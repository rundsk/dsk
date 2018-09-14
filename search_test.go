// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/blevesearch/bleve/analysis/lang/de"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/go-yaml/yaml"
)

func TestFullSearchGermanWordPartials(t *testing.T) {
	contents := `# Farben

	Blau, grün, gelb, violett sie sind wunderschön.
	Nur rot mag ich nicht gerne.
	`

	tmp, s := setupSearchTest(t, "de", "Farben", contents)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("fa", true)
	expectFullSearchResult(t, rs, "Farben")

	rs, _, _, _, _ = s.FullSearch("farbe", true)
	expectFullSearchResult(t, rs, "Farben")

	rs, _, _, _, _ = s.FullSearch("farben", true)
	expectFullSearchResult(t, rs, "Farben")
}

func TestFullSearchTitleUmlauts(t *testing.T) {
	tmp, s := setupSearchTest(t, "de", "Diversität", "")
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversit", true)
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversität", true)
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestFilterSearchTitleUmlauts(t *testing.T) {
	tmp, s := setupSearchTest(t, "de", "Diversität", "")
	defer teardownSearchTest(tmp, s)

	rs, _, _, _ := s.FilterSearch("Diversit", false)
	expectFilterSearchResult(t, rs, "Diversitat")

	rs, _, _, _ = s.FilterSearch("Diversität", false)
	expectFilterSearchResult(t, rs, "Diversitat")
}

// This sort of test is less useful than a test ranking two candidate matches would be.
func TestFilterSearchWithFuzziness(t *testing.T) {
	tmp, s := setupSearchTest(t, "de", "Diversität", "")
	defer teardownSearchTest(tmp, s)

	// Diversit mangled to Divi
	rs, _, _, _ := s.FilterSearch("Divi", true)
	expectFilterSearchResult(t, rs, "Diversitat")

	// Diversität mangled to Divsersität
	rs, _, _, _ = s.FilterSearch("Diversität", true)
	expectFilterSearchResult(t, rs, "Diversitat")
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

func setupSearchTest(t *testing.T, lang string, name string, contents string) (string, *Search) {
	t.Helper()

	tmp, _ := ioutil.TempDir("", "tree")

	node0 := filepath.Join(tmp, name)
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "doc0.md")
	ioutil.WriteFile(doc0, []byte(contents), 0666)

	foo := &Node{root: tmp, path: node0}

	s := &Search{
		getNode: func(url string) (bool, *Node, error) {
			return true, foo, nil
		},
		getAllNodes: func() []*Node {
			ns := make([]*Node, 0)
			ns = append(ns, foo)
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
	return tmp, s
}

func teardownSearchTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
}

func setupSearchScoringTest(t *testing.T, testFile string) (*NodeTree, *Search, map[string]string) {
	t.Helper()

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
	t.Logf("Searching for expected result '%s'...", url)

	for _, hit := range hits {
		t.Logf("...having result '%s'", hit.Node.URL())

		if hit.Node.URL() == url {
			t.Logf("Found result '%s'!", url)
			return
		}
	}
	t.Errorf("Expected '%s' not included in results", url)
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
	t.Logf("Searching for expected result '%s'..", url)

	for _, node := range nodes {
		t.Logf("...having result '%s'", node.URL())

		if node.URL() == url {
			t.Logf("Found result '%s'!", url)
			return
		}
	}
	t.Errorf("Expected '%s' not included in results", url)
}

func keysInOrder(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
