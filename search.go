// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
	"github.com/fatih/color"
)

var (
	// AvailableSearchLangs are languages mapped to their analyzer names.
	AvailableSearchLangs = map[string]string{
		"de": de.AnalyzerName,
		"en": en.AnalyzerName,
	}
)

// NewSearch constructs and initializes a Search. The selected
// languages are validated and checked for availability.
func NewSearch(t *NodeTree, b *MessageBroker, langs []string) (*Search, error) {
	s := &Search{
		getNode:     t.Get,
		getAllNodes: t.GetAll,
		getAuthors:  t.GetAuthors,
		broker:      b,
		isStale:     true,
		done:        make(chan bool),
	}

	for _, l := range langs {
		_, ok := AvailableSearchLangs[l]
		if !ok {
			return s, fmt.Errorf("Unsupported language: %s", l)
		}
	}
	s.langs = langs

	wideIndex, narrowIndex, err := NewIndexes(langs)
	if err != nil {
		return s, err
	}
	s.wideIndex = wideIndex
	s.narrowIndex = narrowIndex

	return s, nil
}

func NewIndexes(langs []string) (bleve.Index, bleve.Index, error) {
	wideIndex, wideErr := bleve.NewMemOnly(NewSearchMapping(langs, true))
	narrowIndex, narrowErr := bleve.NewMemOnly(NewSearchMapping(langs, false))

	if wideErr != nil {
		return wideIndex, narrowIndex, wideErr
	}
	return wideIndex, narrowIndex, narrowErr
}

func NewSearchMapping(langs []string, isWide bool) *mapping.IndexMappingImpl {
	im := bleve.NewIndexMapping()

	if len(langs) > 0 {
		im.DefaultAnalyzer = AvailableSearchLangs[langs[0]]
	}

	sm := bleve.NewTextFieldMapping()
	sm.Analyzer = simple.Name

	km := bleve.NewTextFieldMapping()
	km.Analyzer = keyword.Name

	var tms []*mapping.FieldMapping
	for _, l := range langs {
		tm := bleve.NewTextFieldMapping()
		tm.Analyzer = AvailableSearchLangs[l]
		tms = append(tms, tm)
	}
	node := bleve.NewDocumentMapping()
	node.DefaultAnalyzer = im.DefaultAnalyzer

	node.AddFieldMappingsAt("Titles", tms...)
	node.AddFieldMappingsAt("Tags", sm, km)
	if isWide {
		node.AddFieldMappingsAt("Authors", sm)
		node.AddFieldMappingsAt("Description", tms...)
		node.AddFieldMappingsAt("Docs", tms...)
		node.AddFieldMappingsAt("Files", sm)
		node.AddFieldMappingsAt("Version", sm, km)
		node.AddFieldMappingsAt("Custom", sm)
	}

	im.AddDocumentMapping("article", node)
	return im
}

// Search provides facilities for differnt kinds of search purposes.
// It allows to perform full text searching through FullSearch(),
// wrapping a bleve search index and to perform filtering through
// FilterSearch().
//
// Fuzzy mode can be enabled on a per query basis for FullSearch
// and FilterSearch. The mode should be used if the result
// set doesn't seem large enough. It follows the "It's better
// to have false positives than false negatives" principle:
// https://en.wikipedia.org/wiki/Precision_and_recall
type Search struct {
	sync.RWMutex

	getNode     NodeGetter
	getAllNodes NodesGetter
	getAuthors  func() *Authors

	// Languages that should be used in our mapping/analyzer setup.
	// The first language provided will be used as the default language.
	langs []string

	wideIndex   bleve.Index
	narrowIndex bleve.Index

	// A freshness flag whether the current index is stale and does
	// reflect recent changes from the tree.
	isStale bool

	// Allows to listen for tree change messages.
	broker *MessageBroker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

type FullSearchHit struct {
	Node      *Node
	Fragments []string
}

// StartIndexer installs a go routine ("the indexer") that will
// continously watch for changes to the node tree and will reindex the
// tree if necessary. The indexer can be stopped by sending true into
// Search.done. It'll automatically stop if it detects the broker to
// be closed.
func (s *Search) StartIndexer() {
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	go func() {
		id, messages := s.broker.Subscribe()

		for {
			select {
			case m, ok := <-messages:
				if !ok {
					// Channel is now closed.
					log.Print("Stopping indexer (channel closed)...")
					s.broker.Unsubscribe(id)
					return
				}
				if m.(*Message).typ == MessageTypeTreeChanged {
					s.Lock()
					s.isStale = true
					s.Unlock()
					continue
				}
				// React only on Synced not Loaded messages, the
				// initial indexing and load is triggered manually.
				if m.(*Message).typ == MessageTypeTreeSynced {
					s.Lock()
					// Throw away previous indexes and start from scratch until we
					// have the needs to incrementally invalidate and re-index.
					wideIndex, narrowIndex, err := NewIndexes(s.langs)
					if err != nil {
						s.Unlock()
						log.Print(red.Sprintf("Stopping indexer, failed to construct new indexes: %s", err))
						return
					}
					s.wideIndex = wideIndex
					s.narrowIndex = narrowIndex
					s.Unlock()

					if err := s.IndexTree(); err != nil {
						log.Print(yellow.Sprintf("Failed to index tree: %s", err))
						continue
					}
				}
			case <-s.done:
				log.Print("Stopping indexer (received quit)...")
				s.broker.Unsubscribe(id)
				return
			}
		}
	}()
}

func (s *Search) StopIndexer() {
	s.done <- true
}

func (s *Search) Close() error {
	wideErr := s.wideIndex.Close()
	narrowErr := s.narrowIndex.Close()

	if wideErr != nil {
		return wideErr
	}
	return narrowErr
}

func (s *Search) IndexTree() error {
	start := time.Now()

	wideBatch := s.wideIndex.NewBatch()
	narrowBatch := s.narrowIndex.NewBatch()

	for _, n := range s.getAllNodes() {
		if err := s.IndexNode(n, wideBatch, narrowBatch); err != nil {
			return err
		}
	}

	var err error
	err = s.wideIndex.Batch(wideBatch)
	if err != nil {
		return err
	}
	err = s.narrowIndex.Batch(narrowBatch)
	if err != nil {
		return err
	}

	took := time.Since(start)

	log.Printf("Indexed tree for search in %s", took)
	s.Lock()
	s.isStale = false
	s.Unlock()
	return nil
}

func (s *Search) IndexNode(n *Node, wideBatch, narrowBatch *bleve.Batch) error {
	var as []string
	var ts []string
	var fs []string
	var titles []string

	fs = append(fs, n.Name())
	titles = append(titles, n.Title())

	for _, a := range n.Authors(s.getAuthors()) {
		as = append(as, a.Name)
		as = append(as, a.Email)
	}

	docs, err := n.Docs()
	if err != nil {
		return err
	}
	for _, doc := range docs {
		text, err := doc.CleanText()
		if err != nil {
			return err
		}
		ts = append(ts, string(text))
		fs = append(fs, doc.Name())
		titles = append(titles, doc.Title())
	}

	assets, err := n.Assets()
	if err != nil {
		return err
	}
	for _, a := range assets {
		fs = append(fs, a.Name())
		titles = append(titles, a.Title())
	}

	wideData := struct {
		Authors     []string
		Description string
		Docs        []string
		Files       []string
		Tags        []string
		Titles      []string
		Version     string
		Custom      interface{}
	}{
		Authors:     as,
		Description: n.Description(),
		Docs:        ts,
		Files:       fs,
		Tags:        n.Tags(),
		Titles:      titles,
		Version:     n.Version(),
		Custom:      n.Custom(),
	}
	narrowData := struct {
		Tags   []string
		Titles []string
	}{
		Tags:   wideData.Tags,
		Titles: wideData.Titles,
	}

	s.RLock()
	wideBatch.Index(n.URL(), wideData)
	narrowBatch.Index(n.URL(), narrowData)
	s.RUnlock()

	for _, v := range n.Children {
		s.IndexNode(v, wideBatch, narrowBatch)
	}
	return nil
}

// FullSearch performs a full text search over all possible attributes
// of each node using the wide index. Returns a slice of FullSearchHits.
func (s *Search) FullSearch(q string) ([]*FullSearchHit, int, time.Duration, bool, error) {
	s.RLock()
	defer s.RUnlock()

	mq := bleve.NewMatchQuery(q)
	mq.SetFuzziness(2)
	req := bleve.NewSearchRequest(mq)

	res, err := s.wideIndex.Search(req)
	if err != nil {
		return nil, 0, time.Duration(0), s.isStale, fmt.Errorf("Query '%s' failed: %s", q, err)
	}

	hits := make([]*FullSearchHit, 0, len(res.Hits))
	for _, hit := range res.Hits {
		ok, n, err := s.getNode(hit.ID)
		if err != nil {
			return hits, int(res.Total), res.Took, s.isStale, fmt.Errorf("Failed to get node for hit %s: %s", hit.ID, err)
		}
		if !ok {
			log.Printf("Node for hit %s not found, skipping hit", hit.ID)
			continue
		}

		fragments := make([]string, 0)

		// We only want fragments from the description or the docs, not title or the files
		for _, subFragment := range hit.Fragments["Description"] {
			fragments = append(fragments, subFragment)
		}

		for _, subFragment := range hit.Fragments["Docs"] {
			fragments = append(fragments, subFragment)
		}

		hits = append(hits, &FullSearchHit{n, fragments})
	}
	return hits, int(res.Total), res.Took, s.isStale, nil
}

// FilterSearch performs a narrow restricted prefix search on the
// node's visible attributes (the title) plus tags using the narrow
// index by default. Returns a slice of found unique Nodes.
func (s *Search) FilterSearch(q string, useWideIndex bool) ([]*Node, int, time.Duration, bool, error) {
	s.RLock()
	defer s.RUnlock()

	queryStrings := strings.Split(q, " ")
	var pqs []query.Query

	for _, queryString := range queryStrings {
		pq := bleve.NewPrefixQuery(queryString)
		pqs = append(pqs, pq)
	}

	cq := bleve.NewConjunctionQuery(pqs...)

	req := bleve.NewSearchRequest(cq)

	var res *bleve.SearchResult
	var err error
	if useWideIndex {
		res, err = s.wideIndex.Search(req)
	} else {
		res, err = s.narrowIndex.Search(req)
	}
	if err != nil {
		return nil, 0, time.Duration(0), s.isStale, fmt.Errorf("Query '%s' failed: %s", q, err)
	}

	var nodes []*Node
	seen := make(map[string]bool)
	for _, hit := range res.Hits {
		ok, n, err := s.getNode(hit.ID)
		if err != nil {
			return nodes, len(nodes), res.Took, s.isStale, fmt.Errorf("Failed to get node for hit %s: %s", hit.ID, err)
		}
		if _, hasSeen := seen[n.URL()]; hasSeen {
			continue // Keep nodes unique.
		}
		if !ok {
			log.Printf("Node for hit %s not found, skipping hit", hit.ID)
			continue
		}
		nodes = append(nodes, n)
		seen[n.URL()] = true
	}
	return nodes, len(nodes), res.Took, s.isStale, nil
}

// LegacyFilterSearch performs a narrow restricted haystack/needle
// search on the node's visible attributes (the title) plus tags &
// keywords.
//
// A new filter search has been introduced for APIv2, which we can't
// simply switch into a APIv1 backwards compatible maintaining mode.
func (s *Search) LegacyFilterSearch(q string) ([]*Node, int, time.Duration, error) {
	start := time.Now()

	var results []*Node

	matches := func(source string, target string) bool {
		if source == "" {
			return false
		}
		return strings.Contains(strings.ToLower(target), strings.ToLower(source))
	}

Outer:
	for _, n := range s.getAllNodes() {
		if matches(q, n.Title()) {
			results = append(results, n)
			continue Outer
		}
		for _, v := range n.Tags() {
			if matches(q, v) {
				results = append(results, n)
				continue Outer
			}
		}
		for _, v := range n.Keywords() {
			if matches(q, v) {
				results = append(results, n)
				continue Outer
			}
		}
	}
	return results, len(results), time.Since(start), nil
}
