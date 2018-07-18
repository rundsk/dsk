package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/mapping"
)

func createIndex() (bleve.Index, error) {
	indexMapping := bleve.NewIndexMapping()
	addMappings(indexMapping)
	memIndex, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, err
	}

	return memIndex, nil
}

// AddMappings attempts to be semi general purpose, and includes both
// a tiny bit of fuzzing and exact matches.
func addMappings(indexMapping *mapping.IndexMappingImpl) {
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
}
