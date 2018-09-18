// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/go-yaml/yaml"
)

func TestFullSearchFindsFullWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()

	s := setupSearchTest(t, tmp, "de", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversität", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversität", false)
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestFuzzyFullSearchWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("col", true)
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("color", true)
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("color", true)
	expectFullSearchResult(t, rs, "Colors")
}

func TestFuzzyFullSearchGermanWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversit", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversita", true)
	expectFullSearchResult(t, rs, "Diversitat")
}

// This is the inversion of TestFullSearchGermanWordPartials
func TestOnlyFuzzyModeFindsPartialWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversit", false)
	expectNoFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversita", false)
	expectNoFullSearchResult(t, rs, "Diversitat")
}

func TestGermanFullSearchNormalizesUmlauts(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversität", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitat", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitaet", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversität", false)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitat", false)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitaet", false)
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestEnglishFullSearchDoesNotNormalizeUmlauts(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	// Exact matches always work, independent of languages.
	rs, _, _, _, _ := s.FullSearch("Diversität", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitat", true)
	expectFullSearchResult(t, rs, "Diversitat")

	// Exact matches always work, independent of languages.
	rs, _, _, _, _ = s.FullSearch("Diversität", false)
	expectFullSearchResult(t, rs, "Diversitat")

	// Cannot normalize Umlauts
	rs, _, _, _, _ = s.FullSearch("Diversitat", false)
	expectNoFullSearchResult(t, rs, "Diversitat")
}

func TestConsidersStopwords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "The Diversity"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("The", true)
	expectNoFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("The", false)
	expectNoFullSearchResult(t, rs, "Diversity")
}

func TestSearchFindsAllTagsWhenProvidedAsSlice(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"foo", "bar"},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo", false)
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar", false)
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersMultipleDocs(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("0.md", []byte("lorem ipsum foo"))
	n.CreateDoc("1.md", []byte("dolor amet bar"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo", false)
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar", false)
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document.md", false)
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersTitles(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", n)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document", false)
	expectFullSearchResult(t, rs, "Diversity")
}

func TestTruePositiveFullSearchScore(t *testing.T) {
	const scoreThreshold = 0.8

	tr, s, tests := setupSearchScoringTest(t, "./test/true_positives_fuzzy_search_score.yaml")
	defer teardownSearchScoringTest(tr, s)

	// Avoid division by zero errors at the cost of a bit of precision.
	succeeded := 1
	testCount := 1
	for _, query := range keysInOrder(tests) {
		shouldBeIn := tests[query]
		hits, _, _, _, _ := s.FullSearch(query, true)

		if hasFullSearchResult(t, hits, shouldBeIn) {
			succeeded++

			if len(hits) > 1 {
				foundPaths := []string{}

				for _, hit := range hits {
					if hit.Node.URL() != shouldBeIn {
						foundPaths = append(foundPaths, hit.Node.URL())
					}
				}
				t.Logf("Query found with extras:\nquery: %s\nexpected: %v\nextras: %v", query, shouldBeIn, foundPaths)
			}
		} else {
			foundPaths := []string{}

			for _, hit := range hits {
				foundPaths = append(foundPaths, hit.Node.URL())
			}
			t.Logf("Query result not found:\nquery: %s\nactual: %v\nexpected: %v", query, foundPaths, shouldBeIn)
		}
		testCount++
	}

	truePositive := float64(succeeded) / float64(testCount)

	if truePositive < scoreThreshold {
		t.Errorf("True positive search scoring on test/design_system was %.2f (min required is %.2f)", truePositive, scoreThreshold)
	}
}

func setupSearchTest(t *testing.T, tmp string, lang string, node *Node) *Search {
	t.Helper()
	log.SetOutput(ioutil.Discard)

	s := &Search{
		getNode: func(url string) (bool, *Node, error) {
			return true, node, nil
		},
		getAllNodes: func() []*Node {
			ns := make([]*Node, 0)
			ns = append(ns, node)
			return ns
		},
		getAuthors: func() *Authors {
			a := &Authors{}
			a.Add(&Author{Name: "Randall Hyman", Email: "randall@evilcorp.org"})
			return a
		},
		available: map[string]string{
			"de": de.AnalyzerName,
			"en": en.AnalyzerName,
		},
		langs:  []string{lang},
		broker: NewMessageBroker(), // Allow to mount indexer, and to Close()
		done:   make(chan bool),    // Do not block on Close()
	}
	if err := s.Open(); err != nil {
		t.Fatal(err)
	}
	s.IndexTree()
	return s
}

func teardownSearchTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
}

func setupSearchScoringTest(t *testing.T, testFile string) (*NodeTree, *Search, map[string]string) {
	t.Helper()
	log.SetOutput(ioutil.Discard)

	// Do not initialize watcher and broker, we only need
	// them to fullfill the interface.
	w := NewWatcher("test/design_system")
	b := NewMessageBroker()

	tr := NewNodeTree("test/design_system", w, b)
	tr.Open()
	tr.Sync()

	s := NewSearch(tr, b, []string{"en", "de"})
	s.Open()
	s.IndexTree()

	var tests map[string]string
	raw, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Unable to read scoring test file: %s", err)
	}
	if err := yaml.Unmarshal(raw, &tests); err != nil {
		t.Fatalf("Unable to deserialize scoring test file: %s", err)
	}

	return tr, s, tests
}

func teardownSearchScoringTest(tr *NodeTree, s *Search) {
	s.Close()
	tr.Close()
}

func expectFullSearchResult(t *testing.T, hits []*SearchHit, url string) {
	t.Helper()

	for _, hit := range hits {
		if hit.Node.URL() == url {
			return
		}
	}
	t.Errorf("Expected '%s', but not included in results", url)
}

func expectNoFullSearchResult(t *testing.T, hits []*SearchHit, url string) {
	t.Helper()

	for _, hit := range hits {
		if hit.Node.URL() == url {
			t.Errorf("Not expected '%s' to be included in results", url)
		}
	}
}

func hasFullSearchResult(t *testing.T, hits []*SearchHit, url string) bool {
	t.Helper()

	for _, hit := range hits {
		if hit.Node.URL() == url {
			return true
		}
	}
	return false
}

func expectFilterSearchResult(t *testing.T, nodes []*Node, url string) {
	t.Helper()

	for _, node := range nodes {
		if node.URL() == url {
			return
		}
	}
	t.Errorf("Expected '%s', but not included in results", url)
}

func expectNoFilterSearchResult(t *testing.T, nodes []*Node, url string) {
	t.Helper()

	for _, node := range nodes {
		if node.URL() == url {
			t.Errorf("Not expected '%s' to be included in results", url)
		}
	}
}

func keysInOrder(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
