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

	"github.com/go-yaml/yaml"
)

func TestTruePositiveFullSearchScore(t *testing.T) {
	keysInOrder := func(m map[string]string) []string {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		return keys
	}
	const scoreThreshold = 0.8

	tmp, _ := ioutil.TempDir("", "tree")

	createSearchScoringDesignSystem(t, tmp)
	tests := readSearchScoringTests(t, "./test/true_positives_fuzzy_search_score.yaml")

	s := setupSearchScoringTest(t, tmp)
	defer teardownSearchScoringTest(tmp, s)

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

func createSearchScoringDesignSystem(t *testing.T, tmp string) {
	var n *Node

	n = NewNode(filepath.Join(tmp, "Style"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Components"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Components/Displaying Content"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Components/Displaying Content/List"), tmp)
	n.Create()
	n.CreateDoc("About.md", []byte("The following visual design has been agreed upon by our team:"))
	n.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"progress/draft"},
	})

	n = NewNode(filepath.Join(tmp, "Components/Displaying Content/Input"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Components/Displaying Content/Navigation"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Content"), tmp)
	n.Create()

	n = NewNode(filepath.Join(tmp, "Content/Glossary"), tmp)
	n.Create()
	n.CreateMeta("meta.yaml", &NodeMeta{
		Authors:     []string{"randall@evilcorp.org"},
		Tags:        []string{"Language"},
		Description: "The things we actually mean when we say words that sound fancy.",
		Version:     "2",
	})

	n = NewNode(filepath.Join(tmp, "Content/Tone of Voice"), tmp)
	n.Create()
	n.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"Language"},
	})

	n = NewNode(filepath.Join(tmp, "Culture"), tmp)
	n.Create()
}

func readSearchScoringTests(t *testing.T, path string) map[string]string {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Unable to read scoring test file: %s", err)
	}
	var tests map[string]string

	if err := yaml.Unmarshal(raw, &tests); err != nil {
		t.Fatalf("Unable to deserialize scoring test file: %s", err)
	}
	return tests
}

func setupSearchScoringTest(t *testing.T, tmp string) *Search {
	t.Helper()
	log.SetOutput(ioutil.Discard)

	// Do not initialize watcher and broker, we only need
	// them to fullfill the constructor signature.
	w := NewWatcher(tmp)
	b := NewMessageBroker()
	as := NewAuthors(tmp)

	tr := NewNodeTree(tmp, as, nil, w, b)
	tr.Sync()

	s, _ := NewSearch(tr, b, []string{"en"})
	s.IndexTree()

	return s
}

func teardownSearchScoringTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
	log.SetOutput(os.Stderr)
}
