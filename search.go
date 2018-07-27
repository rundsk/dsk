// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/mapping"
)

func NewSearch(t *NodeTree, b *MessageBroker, langs []string) *Search {
	return &Search{
		getNode:     t.Get,
		getAllNodes: t.GetAll,
		langs:       langs,
		broker:      b,
		done:        make(chan bool),
	}
}

// Search wraps a bleve search index and can be queried for results.
type Search struct {
	getNode     NodeGetter
	getAllNodes NodesGetter

	// Languages we support in our mapping/analyzer setup.
	langs []string

	index bleve.Index

	// Allows to listen for tree change messages.
	broker *MessageBroker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

// Open installs a go routine ("the indexer") that will continously
// watch for changes to the node tree and will reindex the tree
// if necessary. The indexer can be stopped by sending true into
// Search.done. It'll automatically stop if it detects the broker to
// be closed.
func (s *Search) Open() error {
	memIndex, err := bleve.NewMemOnly(s.mapping())
	s.index = memIndex

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
				if m.(*Message).typ == MessageTypeTreeSynced {
					s.IndexTree()
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
	log.Printf("Populating search index from tree...")

	for _, n := range s.getAllNodes() {
		if err := s.IndexNode(n); err != nil {
			return err
		}
	}
	took := time.Since(start)

	log.Printf("Indexed tree for search in %s", took)
	return nil
}

func (s *Search) IndexNode(n *Node) error {
	n.Lock()
	defer n.Unlock()

	dirEntries, err := n.Docs()
	if err != nil {
		return err
	}

	text := []string{}

	for _, nDoc := range dirEntries {
		fileName := nDoc.path
		switch filepath.Ext(fileName) {
		// This explicitly does not convert the .md to HTML
		// with the view that signal to noise is lower in .md than HTML
		case ".md":
			rawBytes, err := nDoc.Raw()
			if err != nil {
				return err
			}

			stringified := string(rawBytes[:len(rawBytes)])
			text = append(text, stringified)
			// TODO: Index other the parts of the node:
			// - assets: read exif data?
		}
	}

	data := struct {
		Text      string
		FileNames []string
		Path      string
	}{
		Text:      strings.Join(text, "\n\n"),
		FileNames: nil,
		Path:      n.URL(),
	}

	s.index.Index(n.URL(), data)

	for _, v := range n.Children {
		s.IndexNode(v)
	}
	return nil
}

// FullSearch is a superset of NarrowSearch in that it performs a
// search over all possible attributes of each node. It does behave
// more like a usual search people are used to.
func (s *Search) FullSearch(query string) ([]*Node, int, time.Duration) {
	start := time.Now()

	mq := bleve.NewMatchQuery(query)
	mq.SetFuzziness(2)
	disjunctionQuery := bleve.NewDisjunctionQuery(mq, bleve.NewTermQuery(query), bleve.NewPrefixQuery(query))

	bSearch := bleve.NewSearchRequest(disjunctionQuery)
	searchResults, err := s.index.Search(bSearch)
	if err != nil {
		log.Fatalf("Query: '%s' failed...", query)
	}

	var results []*Node
	for _, hit := range searchResults.Hits {
		ok, node, err := s.getNode(hit.ID)
		if !ok || err != nil {
			log.Fatalf("For hit %s (ok? %t) something went wrong\n%s", hit.ID, ok, err)
		}
		results = append(results, node)
	}

	return results, len(results), time.Since(start)
}

// FilterSearch performs a narrow restricted fuzzy and term search on
// the node's visible attributes (the title) plus tags & keywords.
//
// We dealt with results where certain things that should have matched
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
// What is used by bleve for fuzzy matching under the hood, Levenstein
// distances weren't enough and weren't able to match this on its own.
// Especially for just a few characters typed.
//
// "It's better to have false positives than false negatives"
// https://en.wikipedia.org/wiki/Precision_and_recall
func (s *Search) FilterSearch(query string) ([]*Node, int, time.Duration) {
	mq := bleve.NewMatchQuery(query)
	mq.SetFuzziness(2)
	disjunctionQuery := bleve.NewDisjunctionQuery(mq, bleve.NewTermQuery(query), bleve.NewPrefixQuery(query))

	bSearch := bleve.NewSearchRequest(disjunctionQuery)
	searchResults, err := s.index.Search(bSearch)
	if err != nil {
		log.Fatalf("Query: '%s' failed...", query)
	}

	results := make([]*Node, 0, searchResults.Total)
	for _, hit := range searchResults.Hits {
		ok, node, err := s.getNode(hit.ID)
		if !ok || err != nil {
			log.Fatalf("For hit %s (ok? %t) something went wrong\n%s", hit.ID, ok, err)
		}
		results = append(results, node)
	}

	return results, int(searchResults.Total), searchResults.Took
}

// TODO: Add language specific analyzers by looking at s.langs
func (s *Search) mapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()

	node := bleve.NewDocumentMapping()
	node.DefaultAnalyzer = de.AnalyzerName
	germanTextMapping := bleve.NewTextFieldMapping()
	germanTextMapping.Analyzer = de.AnalyzerName
	englishTextMapping := bleve.NewTextFieldMapping()
	englishTextMapping.Analyzer = "en"
	node.AddFieldMappingsAt("Text", germanTextMapping, englishTextMapping)

	fileNamesTextFieldMapping := bleve.NewTextFieldMapping()
	fileNamesTextFieldMapping.Analyzer = simple.Name
	node.AddFieldMappingsAt("FileNames", fileNamesTextFieldMapping, germanTextMapping, englishTextMapping)

	pathMapping := bleve.NewTextFieldMapping()
	pathMapping.Analyzer = keyword.Name
	node.AddFieldMappingsAt("Path", pathMapping, germanTextMapping, englishTextMapping)

	// Whether or not mappings work correctly with arrays remains to be seen.
	// Stemming certainly won't for fields like authors
	authorMapping := bleve.NewTextFieldMapping()
	authorMapping.Analyzer = keyword.Name
	node.AddFieldMappingsAt("Authors", authorMapping)

	descriptionMapping := bleve.NewTextFieldMapping()
	descriptionMapping.Analyzer = "en"
	descriptionKeywordMapping := bleve.NewTextFieldMapping()
	descriptionKeywordMapping.Analyzer = keyword.Name
	node.AddFieldMappingsAt("Description", descriptionMapping, descriptionKeywordMapping)

	tagMapping := bleve.NewTextFieldMapping()
	tagMapping.Analyzer = keyword.Name
	node.AddFieldMappingsAt("Tags", tagMapping)

	versionMapping := bleve.NewTextFieldMapping()
	versionMapping.Analyzer = keyword.Name
	node.AddFieldMappingsAt("Version")

	indexMapping.AddDocumentMapping("article", node)

	return indexMapping
}
