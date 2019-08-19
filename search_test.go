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
	"testing"
)

// Tests for FullSearch:

func TestFullSearchAuthorsEmail(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &NodeMeta{
		Authors: []string{"randall@evilcorp.org"},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("randall@evilcorp.org")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("randall")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("evilcorp.org")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("evilcrp.org")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchVersion(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &NodeMeta{
		Version: "2",
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("2")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("Version:2")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchDescription(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &NodeMeta{
		Description: "The things we actually mean when we say words that sound fancy.",
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("fancy")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchCustom(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &NodeMeta{
		Custom: map[string][]string{
			"Synomyms": []string{
				"foo",
				"bar",
			},
		},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchDocumentContents(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()
	n.CreateDoc("About.md", []byte("The following visual design has been agreed upon by our team:"))

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("visual design")
	expectFullSearchResult(t, rs, "Colors")
}

// Tests for FilterSearch:

func TestFilterSearchPrefixes(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("c", false)
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("co", false)
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("col", false)
	expectFilterSearchResult(t, rs, "Colors")
}

func TestFilterSearchWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Navigation"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("na", false)
	expectFilterSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FilterSearch("naviga", false)
	expectFilterSearchResult(t, rs, "Navigation")
}

func TestFilterSearchGermanWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("Diversit", false)
	expectFilterSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FilterSearch("Diversita", false)
	expectFilterSearchResult(t, rs, "Diversitat")
}

// Tests for search in general, often testing only FullSearch, but
// assumes FilterSearch behaves the same, as both use the same search
// backend:
func TestSearchWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Navigation"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("na")
	expectFullSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FullSearch("nav")
	expectFullSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FullSearch("naviga")
	expectFullSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FullSearch("navigati")
	expectFullSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FullSearch("navigatio")
	expectFullSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FullSearch("navigation")
	expectFullSearchResult(t, rs, "Navigation")
}

func TestSearchFindsFullWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()

	s := setupSearchTest(t, tmp, "de", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversität")
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversität")
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestSearchFindsMultipleWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := NewNode(filepath.Join(tmp, "Great"), tmp)
	n0.Create()

	n1 := NewNode(filepath.Join(tmp, "Fantastic"), tmp)
	n1.Create()

	n2 := NewNode(filepath.Join(tmp, "Amazing"), tmp)
	n2.Create()

	s := setupSearchTest(t, tmp, "de", []*Node{n0, n1, n2})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("fantastic great")
	expectFullSearchResult(t, rs, "Fantastic")
	expectFullSearchResult(t, rs, "Great")
}

func TestGermanSearchNormalizesUmlauts(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("Diversität")
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitat")
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("Diversitaet")
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestExactMatchesWorkIndependentOfLang(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	// Exact matches always work, independent of languages.
	rs, _, _, _, _ := s.FullSearch("Diversität")
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestSearchConsidersStopwords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "The Diversity"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("The")
	expectNoFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("The")
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

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("foo bar")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersMultipleDocs(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("0.md", []byte("lorem ipsum foo"))
	n.CreateDoc("1.md", []byte("dolor amet bar"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document.md")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestSearchConsidersTitles(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := NewNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document")
	expectFullSearchResult(t, rs, "Diversity")
}

// Search test helpers:

func setupSearchTest(t *testing.T, tmp string, lang string, nodes []*Node) *Search {
	t.Helper()
	log.SetOutput(ioutil.Discard)

	wideIndex, narrowIndex, _ := NewIndexes(lang)

	lookup := make(map[string]*Node)
	for _, n := range nodes {
		lookup[n.URL()] = n
	}

	s := &Search{
		getNode: func(url string) (bool, *Node, error) {
			if n, ok := lookup[url]; ok {
				return true, n, nil
			}
			return false, nil, nil
		},
		getAllNodes: func() []*Node {
			ns := make([]*Node, 0)

			for _, n := range lookup {
				ns = append(ns, n)
			}
			return ns
		},
		getAuthors: func() *Authors {
			a := &Authors{}
			a.Add(&Author{Name: "Randall Hyman", Email: "randall@evilcorp.org"})
			return a
		},
		lang:        lang,
		wideIndex:   wideIndex,
		narrowIndex: narrowIndex,
		broker:      NewMessageBroker(), // Allow to mount indexer, and to Close()
		done:        make(chan bool),    // Do not block on Close()
	}
	s.IndexTree()
	return s
}

func teardownSearchTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
	log.SetOutput(os.Stderr)
}

func expectFullSearchResult(t *testing.T, hits []*FullSearchHit, url string) {
	t.Helper()

	for _, hit := range hits {
		if hit.Node.URL() == url {
			t.Logf("Found node URL '%s' in results", url)
			return
		}
	}
	t.Errorf("Expected '%s', but not included in results", url)
}

func expectNoFullSearchResult(t *testing.T, hits []*FullSearchHit, url string) {
	t.Helper()

	for _, hit := range hits {
		if hit.Node.URL() == url {
			t.Errorf("Not expected '%s' to be included in results", url)
		}
	}
}

func hasFullSearchResult(t *testing.T, hits []*FullSearchHit, url string) bool {
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
			t.Logf("Found node URL '%s' in results", url)
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

func TestFilterSearchMultipleTagsWithLogicalAndInQuery(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"foo", "bar", "qux"},
	})
	n0.Load()

	n1 := NewNode(filepath.Join(tmp, "Navigation"), tmp)
	n1.Create()
	n1.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"foo"},
	})
	n1.Load()

	n2 := NewNode(filepath.Join(tmp, "Type"), tmp)
	n2.Create()
	n2.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"bar"},
	})
	n2.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n0, n1, n2})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("foo", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FilterSearch("bar", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo bar", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo col", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo shadows", false)
	expectNoFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")
}
}

func TestFilterSearchNamespacedTags(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := NewNode(filepath.Join(tmp, "Colors"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"foo", "status/draft"},
	})
	n0.Load()

	n1 := NewNode(filepath.Join(tmp, "Navigation"), tmp)
	n1.Create()
	n1.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"status/ready"},
	})
	n1.Load()

	n2 := NewNode(filepath.Join(tmp, "Type"), tmp)
	n2.Create()
	n2.CreateMeta("meta.yaml", &NodeMeta{
		Tags: []string{"draft"},
	})
	n2.Load()

	s := setupSearchTest(t, tmp, "en", []*Node{n0, n1, n2})
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("status", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("statusdraft", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("draft", false)
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectFilterSearchResult(t, rs, "Type")
}
