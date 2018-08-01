// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

const scoreThreshold = 0.8

var (
	here string
)

func TestTruePositiveSearchScore(t *testing.T) {
	s := setupScoringTest()
	defer teardownScoringTest(s)

	// Potentially we want this in a different file (tsv or yaml(?))
	tests := map[string]string{
		"randall":     "Content/Glossary",
		"evilcorp":    "Content/Glossary",
		"word":        "Content/Glossary",
		"wor":         "Content/Glossary",
		"fanc":        "Content/Glossary",
		"fancy":       "Content/Glossary",
		"what we say": "Content/Tone_of_Voice",
	}

	// Avoid division by zero errors at the cost of a bit of precision.
	succeeded := 1
	testCount := 1
	for query, shouldBeIn := range tests {
		rs, _, _ := s.FilterSearch(query)

		if stringInSlice(shouldBeIn, rs) {
			succeeded++
		} else {
			foundPaths := []string{}

			for _, n := range rs {
				foundPaths = append(foundPaths, n.URL())
			}

			log.WithFields(log.Fields{
				"query":    query,
				"actual":   foundPaths,
				"expected": shouldBeIn,
			}).Warn("Query result not found")
		}
		testCount++
	}

	truePositive := float64(succeeded) / float64(testCount)

	log.Printf("True positive search scoring on %s was %.2f (min required is %.2f)", here, truePositive, scoreThreshold)
	if truePositive < scoreThreshold {
		t.Fail()
	}
}

func setupScoringTest() *Search {
	here = "test/design_system" // assignment to global
	w := NewWatcher(here)
	if err := w.Open(IgnoreNodesRegexp); err != nil {
		log.Fatalf("Failed to install watcher: %s", err)
	}
	watcher = w // assign to global

	log.Print("Starting message broker...")
	broker = NewMessageBroker() // assign to global
	broker.Start()

	log.Print("Opening tree...")
	tree = NewNodeTree(here, watcher, broker) // assign to global

	if err := tree.Open(); err != nil {
		log.Fatalf("Failed to open tree: %s", err)
	}
	if err := tree.Sync(); err != nil {
		log.Fatalf("Failed to perform initial tree sync: %s", err)
	}

	log.Print("Opening test search index...")
	search = NewSearch(tree, broker, []string{"en", "de"}) // assign to global
	if err := search.Open(); err != nil {
		log.Fatalf("Failed to open search index: %s", err)
	}

	if err := search.IndexTree(); err != nil {
		log.Fatalf("Failed to perform initial test tree indexing: %s", err)
	}
	return search
}

func teardownScoringTest(s *Search) {
	s.Close()
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
