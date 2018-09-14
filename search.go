// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/fatih/color"
)

func NewSearch(t *NodeTree, b *MessageBroker, langs []string) *Search {
	return &Search{
		getNode:     t.Get,
		getAllNodes: t.GetAll,
		getAuthors:  t.GetAuthors,
		langs:       langs,
		available: map[string]string{
			"de": de.AnalyzerName,
			"en": en.AnalyzerName,
		},
		broker:  b,
		isStale: true,
		done:    make(chan bool),
	}
}

// Search wraps a bleve search index and can be queried for results.
//
// It follows the "It's better to have false positives than
// false negatives" principle:
// https://en.wikipedia.org/wiki/Precision_and_recall
//
// On a per query bassi fuzzy mode can enabled for FullSearch and
// FilterSearch. The mode should be used if the result set doesn't
// seem large enough.
type Search struct {
	getNode     NodeGetter
	getAllNodes NodesGetter
	getAuthors  func() *Authors

	// Languages that should be used in our mapping/analyzer setup.
	// The first language provided will be used as the default language.
	langs []string

	// Available languages mapped to their analyzer names.
	available map[string]string

	index bleve.Index

	// A freshness flag whether the current index is stale and does
	// reflect recent changes from the tree.
	isStale bool

	// Allows to listen for tree change messages.
	broker *MessageBroker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

type SearchHit struct {
	Node *Node
}

// Open installs a go routine ("the indexer") that will continously
// watch for changes to the node tree and will reindex the tree
// if necessary. The indexer can be stopped by sending true into
// Search.done. It'll automatically stop if it detects the broker to
// be closed.
func (s *Search) Open() error {
	ma, err := s.mapping()
	if err != nil {
		return err
	}

	memIndex, err := bleve.NewMemOnly(ma)
	if err != nil {
		return err
	}
	s.index = memIndex

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
					s.isStale = true
					continue
				}
				if m.(*Message).typ == MessageTypeTreeSynced {
					// Throw away previous index and start from scratch until we
					// have the needs to incrementally invalidate and re-index.
					memIndex, err := bleve.NewMemOnly(ma)
					if err != nil {
						log.Print(red.Sprintf("Stopping indexer, failed to construct new index: %s", err))
						return
					}
					s.index = memIndex

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
	return err
}

func (s *Search) Close() error {
	log.Print("Search index is closing...")
	s.done <- true // Stop indexer
	return s.index.Close()
}

func (s *Search) IndexTree() error {
	start := time.Now()

	for _, n := range s.getAllNodes() {
		if err := s.IndexNode(n); err != nil {
			return err
		}
	}
	took := time.Since(start)

	log.Printf("Indexed tree for search in %s", took)
	s.isStale = false
	return nil
}

func (s *Search) IndexNode(n *Node) error {
	var ab strings.Builder
	var tb strings.Builder

	s.getAuthors()

	authors := n.Authors(s.getAuthors())
	for _, a := range authors {
		ab.WriteString(a.Name)
		ab.WriteString("\n")
		ab.WriteString(a.Email)
		ab.WriteString("\n")
	}

	docs, err := n.Docs()
	if err != nil {
		return err
	}
	for _, doc := range docs {
		text, err := doc.Text()
		if err != nil {
			return err
		}
		tb.WriteString("\n\n")
		tb.Write(text)
	}

	data := struct {
		Authors     string
		Description string
		Docs        string
		Tags        []string
		Title       string
		Version     string
	}{
		Authors:     ab.String(),
		Description: n.Description(),
		Docs:        tb.String(),
		Tags:        n.Tags(),
		Title:       n.Title(),
		Version:     n.Version(),
	}

	s.index.Index(n.URL(), data)

	for _, v := range n.Children {
		s.IndexNode(v)
	}
	return nil
}

// FullSearch performs a full text search over all possible attributes
// of each node.
//
// For fuzzy mode we weren't able to use bleve's Fuzzy query as, we
// dealt with results where certain things that should have matched
// with a raw match query, did not. For example, `Farben` being an
// article we wanted to match, we would type:
//
// | Input        | True positive |
// | -------------|:-------------:|
// | f            | true          |
// | fa           | true          |
// | far          | false         |
// | farb         | false         |
// | farbe        | true          |
// | farben       | true          |
//
// What is used by bleve for fuzzy matching under the hood,
// Levenshtein distances weren't enough and weren't able to match this
// on its own. Especially for just a few characters typed.
func (s *Search) FullSearch(q string, fuzzy bool) ([]*SearchHit, int, time.Duration, bool, error) {
	var req *bleve.SearchRequest

	if fuzzy {
		mq := bleve.NewMatchQuery(q)
		mq.SetFuzziness(2)

		dq := bleve.NewDisjunctionQuery(
			mq,
			bleve.NewTermQuery(q),
			bleve.NewPrefixQuery(q),
		)
		req = bleve.NewSearchRequest(dq)
	} else {
		mq := bleve.NewMatchQuery(q)
		req = bleve.NewSearchRequest(mq)
	}

	res, err := s.index.Search(req)
	if err != nil {
		return nil, 0, time.Duration(0), s.isStale, fmt.Errorf("Query '%s' failed: %s", q, err)
	}

	hits := make([]*SearchHit, 0, len(res.Hits))
	for _, hit := range res.Hits {
		ok, n, err := s.getNode(hit.ID)
		if err != nil {
			return hits, int(res.Total), res.Took, s.isStale, fmt.Errorf("Failed to get node for hit %s: %s", hit.ID, err)
		}
		if !ok {
			log.Printf("Node for hit %s not found, skipping hit", hit.ID)
			continue
		}
		hits = append(hits, &SearchHit{n})
	}
	return hits, int(res.Total), res.Took, s.isStale, nil
}

// FilterSearch performs a narrow restricted prefix search on the
// node's visible attributes (the title) plus tags & keywords.
//
// Does not use search index, as it's not possible to narrow field
// scope on a per query basis. This means we'd need to keep a second
// index just for filter searches. The simplistic approach used her is
// "good enough" to fullfill the requirements.
func (s *Search) FilterSearch(q string, fuzzy bool) ([]*Node, int, time.Duration, error) {
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

func (s *Search) mapping() (*mapping.IndexMappingImpl, error) {
	im := bleve.NewIndexMapping()

	if len(s.langs) > 0 {
		defaultLang := s.langs[0]
		an, ok := s.available[defaultLang]
		if !ok {
			return im, fmt.Errorf("Unsupported language: %s", defaultLang)
		}
		im.DefaultAnalyzer = an
	}

	sm := bleve.NewTextFieldMapping()
	sm.Analyzer = simple.Name

	km := bleve.NewTextFieldMapping()
	km.Analyzer = keyword.Name

	var tms []*mapping.FieldMapping
	for _, l := range s.langs {
		tm := bleve.NewTextFieldMapping()

		an, ok := s.available[l]
		if !ok {
			return im, fmt.Errorf("Unsupported language: %s", l)
		}
		tm.Analyzer = an

		tms = append(tms, tm)
	}
	node := bleve.NewDocumentMapping()
	node.DefaultAnalyzer = im.DefaultAnalyzer

	node.AddFieldMappingsAt("Authors", sm)
	node.AddFieldMappingsAt("Description", tms...)
	node.AddFieldMappingsAt("Docs", tms...)
	node.AddFieldMappingsAt("Tags", sm, km)
	node.AddFieldMappingsAt("Title", tms...)
	node.AddFieldMappingsAt("Version", sm, km)

	im.AddDocumentMapping("article", node)

	return im, nil
}
