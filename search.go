// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/mapping"
)

func NewSearchIndex() *SearchIndex {
	return &SearchIndex{}
}

// SearchIndex wraps bleve search index.
type SearchIndex struct {
	index bleve.Index
}

func (si *SearchIndex) Open() error {
	memIndex, err := bleve.NewMemOnly(si.mapping())
	si.index = memIndex
	return err
}

func (si *SearchIndex) Close() error {
	return si.index.Close()
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
