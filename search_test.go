// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFilterSearchGermanWordPartials(t *testing.T) {
	contents := `# Farben

	Blau, grün, gelb, violett sie sind wunderschön.
	Nur rot mag ich nicht gerne.
	`

	tmp, s := setupDocSearchTest(contents)
	defer teardownSearchTest(tmp, s)

	rs, _, _ := s.FilterSearch("fa")
	expectSearchResult(t, rs, "foo")

	rs, _, _ = s.FilterSearch("farbe")
	expectSearchResult(t, rs, "foo")

	rs, _, _ = s.FilterSearch("farben")
	expectSearchResult(t, rs, "foo")
}

func setupDocSearchTest(contents string) (string, *Search) {
	tmp, _ := ioutil.TempDir("", "tree")

	node0 := filepath.Join(tmp, "foo")
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
		broker: NewMessageBroker(), // Allow to mount indexer, and to Close()
		done:   make(chan bool),    // Do not block on Close()
	}
	s.Open()
	s.IndexTree()
	return tmp, s
}

func teardownSearchTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
}

func expectSearchResult(t *testing.T, rs []*Node, url string) {
	t.Helper()

	for _, r := range rs {
		if r.URL() == url {
			return
		}
	}
	t.Errorf("Expected hit '%s' not included in results", url)
}
