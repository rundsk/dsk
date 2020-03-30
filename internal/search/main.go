// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package search

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
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
	"github.com/rundsk/dsk/internal/ddt"
)

var (
	// AvailableSearchLangs are languages mapped to their analyzer names.
	AvailableSearchLangs = map[string]string{
		"de": de.AnalyzerName,
		"en": en.AnalyzerName,
	}
)

// NewSearch constructs and initializes a Search. The selected
// language is validated and checked for availability.
func NewSearch(path string, t *ddt.Tree, lang string, isPersistent bool) (*Search, error) {
	log.Print("Initializing search...")

	s := &Search{
		path:        path,
		getNode:     t.Get,
		getAllNodes: t.GetAll,
		getTreeHash: t.CalculateHash,
	}

	_, ok := AvailableSearchLangs[lang]
	if !ok {
		return s, fmt.Errorf("unsupported language: %s", lang)
	}
	s.lang = lang

	wideIndex, narrowIndex, err := NewIndexes(s.path, s.lang, isPersistent)
	if err != nil {
		return s, err
	}
	s.wideIndex = wideIndex
	s.narrowIndex = narrowIndex

	go s.IndexTree()
	return s, nil
}

func NewIndexes(path string, lang string, isPersistent bool) (bleve.Index, bleve.Index, error) {
	var wideIndex bleve.Index
	var wideErr error
	widePath := filepath.Join(path, "wide.bleve")
	wideMapping := NewSearchMapping(lang, true)

	var narrowIndex bleve.Index
	var narrowErr error
	narrowPath := filepath.Join(path, "narrow.bleve")
	narrowMapping := NewSearchMapping(lang, false)

	if isPersistent {
		log.Printf("Persisting search indexes in: %s", path)

		wideIndex, wideErr = bleve.New(widePath, wideMapping)
		narrowIndex, narrowErr = bleve.New(narrowPath, narrowMapping)
	} else {
		wideIndex, wideErr = bleve.NewMemOnly(wideMapping)
		narrowIndex, narrowErr = bleve.NewMemOnly(narrowMapping)
	}

	if wideErr != nil {
		return wideIndex, narrowIndex, wideErr
	}
	return wideIndex, narrowIndex, narrowErr
}

func NewSearchMapping(lang string, isWide bool) *mapping.IndexMappingImpl {
	im := bleve.NewIndexMapping()

	sm := bleve.NewTextFieldMapping()
	sm.Analyzer = simple.Name

	km := bleve.NewTextFieldMapping()
	km.Analyzer = keyword.Name

	tm := bleve.NewTextFieldMapping()
	tm.Analyzer = AvailableSearchLangs[lang]

	node := bleve.NewDocumentMapping()

	node.AddFieldMappingsAt("Title", tm)
	node.AddFieldMappingsAt("Tags", km)
	if isWide {
		node.AddFieldMappingsAt("SecondaryTitles", tm)
		node.AddFieldMappingsAt("Authors", sm)
		node.AddFieldMappingsAt("Description", tm)
		node.AddFieldMappingsAt("Docs", tm)
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
type Search struct {
	sync.RWMutex

	// Path to directory holding the indexes, when persistence is
	// activated.
	path string

	getNode     ddt.NodeGetter
	getAllNodes ddt.NodesGetter
	getTreeHash func() (string, error)

	// Language that should be used in our mapping/analyzer setup.
	lang string

	wideIndex   bleve.Index
	narrowIndex bleve.Index

	// Hash of the node tree that was indexed last.
	hash string
}

type FullSearchHit struct {
	Node      *ddt.Node
	Fragments []string
}

func (s *Search) IsStale() bool {
	s.RLock()
	defer s.RUnlock()

	h, _ := s.getTreeHash()
	return s.hash != h
}

func (s *Search) Close() error {
	wideErr := s.wideIndex.Close()
	narrowErr := s.narrowIndex.Close()

	if wideErr != nil {
		return wideErr
	}
	return narrowErr
}

func (s *Search) Refresh() error {
	s.Lock()

	// Throw away previous indexes and start from scratch until we
	// have the needs to incrementally invalidate and re-index.
	wideIndex, narrowIndex, err := NewIndexes("", s.lang, false)
	if err != nil {
		s.Unlock()
		return err
	}
	s.wideIndex = wideIndex
	s.narrowIndex = narrowIndex
	s.Unlock()

	return s.IndexTree()
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

	s.Lock()
	h, _ := s.getTreeHash()
	s.hash = h
	s.Unlock()

	log.Printf("Indexed node tree for search in %s", took)
	return nil
}

func (s *Search) IndexNode(n *ddt.Node, wideBatch, narrowBatch *bleve.Batch) error {
	var as []string
	var ts []string
	var fs []string
	var secondaryTitles []string // Titles of documents and assets.

	fs = append(fs, n.Name())

	for _, a := range n.Authors() {
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
		secondaryTitles = append(secondaryTitles, doc.Title())
	}

	assets, err := n.Assets()
	if err != nil {
		return err
	}
	for _, a := range assets {
		fs = append(fs, a.Name())
		secondaryTitles = append(secondaryTitles, a.Title())
	}

	wideData := struct {
		Authors         []string
		Description     string
		Docs            []string
		Files           []string
		Tags            []string
		Title           string
		SecondaryTitles []string
		Version         string
		Custom          interface{}
	}{
		Authors:         as,
		Description:     n.Description(),
		Docs:            ts,
		Files:           fs,
		Tags:            n.Tags(),
		Title:           n.Title(),
		SecondaryTitles: secondaryTitles,
		Version:         n.Version(),
		Custom:          n.Custom(),
	}
	narrowData := struct {
		Tags  []string
		Title string
	}{
		Tags:  wideData.Tags,
		Title: wideData.Title,
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

	// Prefix query is case sensitive, we want to have it case insensitive.
	qlower := strings.ToLower(q)

	mq := bleve.NewMatchQuery(q)
	mq.SetFuzziness(1)

	pq := bleve.NewPrefixQuery(qlower)

	tmq := bleve.NewMatchQuery(q)
	tmq.SetField("Title")
	tmq.SetBoost(2)

	tpq := bleve.NewPrefixQuery(qlower)
	tpq.SetField("Title")
	tpq.SetBoost(3)

	dq := bleve.NewDisjunctionQuery(
		mq,
		pq,
		tmq,
		tpq,
	)

	req := bleve.NewSearchRequest(dq)
	req.Highlight = bleve.NewHighlight()

	res, err := s.wideIndex.Search(req)
	if err != nil {
		return nil, 0, time.Duration(0), s.IsStale(), fmt.Errorf("query '%s' failed: %s", q, err)
	}

	hits := make([]*FullSearchHit, 0, len(res.Hits))
	for _, hit := range res.Hits {
		ok, n, err := s.getNode(hit.ID)
		if err != nil {
			return hits, int(res.Total), res.Took, s.IsStale(), fmt.Errorf("failed to get node for hit %s: %s", hit.ID, err)
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
	return hits, int(res.Total), res.Took, s.IsStale(), nil
}

// FilterSearch performs a narrow restricted prefix search on the
// node's visible attributes (the title) plus tags using the narrow
// index by default. Returns a slice of found unique Nodes.
func (s *Search) FilterSearch(q string) ([]*ddt.Node, int, time.Duration, bool, error) {
	s.RLock()
	defer s.RUnlock()

	// Prefix query is case sensitive, we want to have it case insensitive.
	qlower := strings.ToLower(q)

	// We split the query for use with prefix query. Tags may contain
	// slashes to namespace tags (1). We want to be able to search
	// each of them. The query may contain (2) spaces, each space
	// separated word should be used as an individual query and AND'ed
	// together.
	splitter := regexp.MustCompile(`[^\p{L}0-9]`)

	var pqs []query.Query
	for _, qlowerPart := range splitter.Split(qlower, -1) {
		pq := bleve.NewPrefixQuery(qlowerPart)
		pqs = append(pqs, pq)
	}

	cq := bleve.NewConjunctionQuery(pqs...)
	req := bleve.NewSearchRequest(cq)
	req.Size = 100

	res, err := s.narrowIndex.Search(req)

	if err != nil {
		return nil, 0, time.Duration(0), s.IsStale(), fmt.Errorf("query '%s' failed: %s", q, err)
	}

	var nodes []*ddt.Node
	seen := make(map[string]bool)
	for _, hit := range res.Hits {
		ok, n, err := s.getNode(hit.ID)
		if err != nil {
			return nodes, len(nodes), res.Took, s.IsStale(), fmt.Errorf("failed to get node for hit %s: %s", hit.ID, err)
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
	return nodes, len(nodes), res.Took, s.IsStale(), nil
}

// LegacyFilterSearch performs a narrow restricted haystack/needle
// search on the node's visible attributes (the title) plus tags &
// keywords.
//
// A new filter search has been introduced for APIv2, which we can't
// simply switch into a APIv1 backwards compatible maintaining mode.
func (s *Search) LegacyFilterSearch(q string) ([]*ddt.Node, int, time.Duration, error) {
	start := time.Now()

	var results []*ddt.Node

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
