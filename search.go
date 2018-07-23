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

func NewSearchIndex(t *NodeTree, b *MessageBroker) *SearchIndex {
	return &SearchIndex{
		tree:   t,
		broker: b,
		done:   make(chan bool),
	}
}

// SearchIndex wraps bleve search index.
type SearchIndex struct {
	tree *NodeTree

	index bleve.Index

	broker *MessageBroker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

func (si *SearchIndex) Open() error {
	memIndex, err := bleve.NewMemOnly(si.mapping())
	si.index = memIndex

	go func() {
		id, messages := si.broker.Subscribe()

		for {
			select {
			case m, ok := <-messages:
				if !ok {
					// Channel is now closed.
					log.Print("Stopping indexer (channel closed)...")
					si.broker.Unsubscribe(id)
					return
				}
				if m.(*Message).typ == MessageTypeTreeSynced {
					si.IndexTree()
				}
			case <-si.done:
				log.Print("Stopping indexer (received quit)...")
				si.broker.Unsubscribe(id)
				return
			}
		}
	}()
	return err
}

func (si *SearchIndex) Close() error {
	log.Print("Search index is closing...")
	si.done <- true // Stop indexer
	return si.index.Close()
}

func (si *SearchIndex) IndexTree() error {
	start := time.Now()
	log.Printf("Populating search index from tree...")

	for _, n := range si.tree.GetAll() {
		if err := si.IndexNode(n); err != nil {
			return err
		}
	}
	took := time.Since(start)

	log.Printf("Indexed tree for search in %s", took)
	return nil
}

func (si *SearchIndex) IndexNode(n *Node) error {
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

	si.Index(n.URL(), data)

	for _, v := range n.Children {
		si.IndexNode(v)
	}
	return nil
}

func (si *SearchIndex) Index(id string, data interface{}) error {
	return si.index.Index(id, data)
}

func (si *SearchIndex) Search(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return si.index.Search(req)
}

// Mapping attempts to be semi general purpose, and includes both
// a tiny bit of fuzzing and exact matches.
//
// TODO: Have english as default and support any additional language,
//       possible configured via a command line option and/or through auto-detection.
func (si *SearchIndex) mapping() *mapping.IndexMappingImpl {
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

	indexMapping.AddDocumentMapping("article", node)

	return indexMapping
}
