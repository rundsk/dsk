// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"sort"
	"testing"

	"github.com/go-yaml/yaml"
)

func TestTruePositiveSearchScore(t *testing.T) {
	scoreThreshold := 0.8

	tr, s, tests := setupSearchScoringTest(t, "./test/true_positives_search_score.yaml")
	defer teardownSearchScoringTest(tr, s)

	// Avoid division by zero errors at the cost of a bit of precision.
	succeeded := 1
	testCount := 1
	for _, query := range keysInOrder(tests) {
		shouldBeIn := tests[query]
		rs, _, _ := s.FullSearch(query)

		if stringInSlice(shouldBeIn, rs) {
			succeeded++

			if len(rs) > 1 {
				foundPaths := []string{}

				for _, n := range rs {
					if n.URL() != shouldBeIn {
						foundPaths = append(foundPaths, n.URL())
					}
				}
				t.Logf("Query found with extras:\nquery: %s\nexpected: %v\nextras: %v", query, shouldBeIn, foundPaths)
			}
		} else {
			foundPaths := []string{}

			for _, n := range rs {
				foundPaths = append(foundPaths, n.URL())
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

// https://stackoverflow.com/a/23332089/1924257
func keysInOrder(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// This feels like it should baked in...
func stringInSlice(a string, list []*Node) bool {
	for _, b := range list {
		if b.URL() == a {
			return true
		}
	}
	return false
}
