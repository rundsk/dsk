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
	log "github.com/sirupsen/logrus"
)

const scoreThreshold = 0.8
const truePositiveFp = "./test/true_positives.yaml"

var (
	here string
)

func TestTruePositiveSearchScore(t *testing.T) {
	s := setupScoringTest()
	defer teardownScoringTest(s)

	raw, err := ioutil.ReadFile(truePositiveFp)
	if err != nil {
		log.Fatalf("Unable to read scoring test file: %s", err)
	}

	var tests map[string]string
	if err := yaml.Unmarshal(raw, &tests); err != nil {
		log.Fatalf("Unable to deserialize scoring test file: %s", err)
	}

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

				log.WithFields(log.Fields{
					"_query":   query,
					"expected": shouldBeIn,
					"extras":   foundPaths,
				}).Debug("Query found with extras")
			}
		} else {
			foundPaths := []string{}

			for _, n := range rs {
				foundPaths = append(foundPaths, n.URL())
			}

			log.WithFields(log.Fields{
				"_query":   query,
				"actual":   foundPaths,
				"expected": shouldBeIn,
			}).Warn("Query result not found")
		}
		testCount++
	}

	truePositive := float64(succeeded) / float64(testCount)

	log.Infof("True positive search scoring on %s was %.2f (min required is %.2f)", here, truePositive, scoreThreshold)
	if truePositive < scoreThreshold {
		t.Fail()
	}
}

func setupScoringTest() *Search {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	here = "test/design_system" // assignment to global
	w := NewWatcher(here)
	if err := w.Open(IgnoreNodesRegexp); err != nil {
		log.Fatalf("Failed to install watcher: %s", err)
	}
	watcher = w // assign to global

	log.Debug("Starting message broker...")
	broker = NewMessageBroker() // assign to global
	broker.Start()

	log.Debug("Opening tree...")
	tree = NewNodeTree(here, watcher, broker) // assign to global

	if err := tree.Open(); err != nil {
		log.Fatalf("Failed to open tree: %s", err)
	}
	if err := tree.Sync(); err != nil {
		log.Fatalf("Failed to perform initial tree sync: %s", err)
	}

	log.Info("Opening test search index...")
	search = NewSearch(tree, broker, []string{"en", "de"}) // assign to global
	if err := search.Open(); err != nil {
		log.Fatalf("Failed to open search index: %s", err)
	}

	if err := search.IndexTree(); err != nil {
		log.Fatalf("Failed to perform initial test tree indexing: %s", err)
	}
	return search
}

func keysInOrder(m map[string]string) []string {
	// https://stackoverflow.com/a/23332089/1924257
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
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
