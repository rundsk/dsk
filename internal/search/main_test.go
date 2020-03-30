// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package search

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/rundsk/dsk/internal/author"
	"github.com/rundsk/dsk/internal/config"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/meta"
)

// Tests for FullSearch:

func TestFullSearchWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Navigation"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
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

func TestFullSearchFindsFullWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()

	s := setupSearchTest(t, tmp, "de", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("diversität")
	expectFullSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FullSearch("diversität")
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestFullSearchUsesLogicalOrWithMultipleWords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Great"), tmp)
	n0.Create()

	n1 := newTestNode(filepath.Join(tmp, "Fantastic"), tmp)
	n1.Create()

	n2 := newTestNode(filepath.Join(tmp, "Amazing"), tmp)
	n2.Create()

	s := setupSearchTest(t, tmp, "de", []*ddt.Node{n0, n1, n2}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("fantastic great")
	expectFullSearchResult(t, rs, "Fantastic")
	expectFullSearchResult(t, rs, "Great")
}

func TestFullSearchExactMatchesWorkIndependentOfLang(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	// Exact matches always work, independent of languages.
	rs, _, _, _, _ := s.FullSearch("diversität")
	expectFullSearchResult(t, rs, "Diversitat")
}

func TestFullSearchConsidersStopwords(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "The Diversity"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("the")
	expectNoFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("the")
	expectNoFullSearchResult(t, rs, "Diversity")
}

func TestFullSearchFindsAllTagsWhenProvidedAsSlice(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo", "bar"},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("foo bar")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestFullSearchConsidersMultipleDocs(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("0.md", []byte("lorem ipsum foo"))
	n.CreateDoc("1.md", []byte("dolor amet bar"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Diversity")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestFullSearchConsidersFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document.md")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestFullSearchConsidersSecondaryTitles(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversity"), tmp)
	n.Create()
	n.CreateDoc("document.md", []byte("lorem ipsum"))
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("document")
	expectFullSearchResult(t, rs, "Diversity")
}

func TestFullSearchAuthorsEmail(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Authors: []string{"randall@evilcorp.org"},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
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

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Version: "2",
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("2")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("Version:2")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchDescription(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Description: "The things we actually mean when we say words that sound fancy.",
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("fancy")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchCustom(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	n.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Custom: map[string][]string{
			"Synomyms": []string{
				"foo",
				"bar",
			},
		},
	})
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("foo")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("bar")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchDocumentContents(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()
	n.CreateDoc("About.md", []byte("The following visual design has been agreed upon by our team:"))

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("visual design")
	expectFullSearchResult(t, rs, "Colors")
}

func TestFullSearchIsCaseInsensitive(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FullSearch("colors")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("Colors")
	expectFullSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FullSearch("coLOrs")
	expectFullSearchResult(t, rs, "Colors")
}

// Tests for FilterSearch:

func TestFilterSearchPrefixes(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("c")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("co")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("col")
	expectFilterSearchResult(t, rs, "Colors")
}

func TestFilterSearchWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Navigation"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("na")
	expectFilterSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FilterSearch("naviga")
	expectFilterSearchResult(t, rs, "Navigation")
}

func TestFilterSearchGermanWordPartials(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Diversität"), tmp)
	n.Create()
	n.Load()

	s := setupSearchTest(t, tmp, "de", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("diversit")
	expectFilterSearchResult(t, rs, "Diversitat")

	rs, _, _, _, _ = s.FilterSearch("diversitä")
	expectFilterSearchResult(t, rs, "Diversitat")
}

func TestFilterSearchTags(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Button"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"react"},
	})
	n0.Load()

	n1 := newTestNode(filepath.Join(tmp, "Form Element"), tmp)
	n1.Create()
	n1.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"react"},
	})
	n1.Load()

	n2 := newTestNode(filepath.Join(tmp, "Radio Button Group"), tmp)
	n2.Create()
	n2.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"react"},
	})
	n2.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n0, n1, n2}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("react")
	expectFilterSearchResult(t, rs, "Button")
	expectFilterSearchResult(t, rs, "Form Element")
	expectFilterSearchResult(t, rs, "Radio Button Group")
}

func TestFilterSearchMultipleTagsWithLogicalAndInQuery(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo", "bar", "qux"},
	})
	n0.Load()

	n1 := newTestNode(filepath.Join(tmp, "Navigation"), tmp)
	n1.Create()
	n1.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo"},
	})
	n1.Load()

	n2 := newTestNode(filepath.Join(tmp, "Type"), tmp)
	n2.Create()
	n2.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"bar"},
	})
	n2.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n0, n1, n2}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("foo")
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Navigation")

	rs, _, _, _, _ = s.FilterSearch("bar")
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo bar")
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo col")
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("foo shadows")
	expectNoFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")
}

func TestFilterSearchIsCaseInsensitive(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n.Create()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("colors")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("Colors")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("coLOrs")
	expectFilterSearchResult(t, rs, "Colors")
}

func TestFilterSearchNamespacedTags(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo", "status/draft"},
	})
	n0.Load()

	n1 := newTestNode(filepath.Join(tmp, "Navigation"), tmp)
	n1.Create()
	n1.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"status/ready"},
	})
	n1.Load()

	n2 := newTestNode(filepath.Join(tmp, "Type"), tmp)
	n2.Create()
	n2.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"draft"},
	})
	n2.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n0, n1, n2}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("status")
	expectFilterSearchResult(t, rs, "Colors")
	expectFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("status/draft")
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectNoFilterSearchResult(t, rs, "Type")

	rs, _, _, _, _ = s.FilterSearch("draft")
	expectFilterSearchResult(t, rs, "Colors")
	expectNoFilterSearchResult(t, rs, "Navigation")
	expectFilterSearchResult(t, rs, "Type")
}

func TestFilterSearchTagsWithSpaces(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Colors"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo", "needs images"},
	})
	n0.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n0}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("needs")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("images")
	expectFilterSearchResult(t, rs, "Colors")

	rs, _, _, _, _ = s.FilterSearch("needs images")
	expectFilterSearchResult(t, rs, "Colors")
}

func TestFilterSearchTagsWithSpacesWhenTitleContainsSpace(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")

	n0 := newTestNode(filepath.Join(tmp, "Color Definition"), tmp)
	n0.Create()
	n0.CreateMeta("meta.yaml", &ddt.NodeMeta{
		Tags: []string{"foo", "needs images"},
	})
	n0.Load()

	s := setupSearchTest(t, tmp, "en", []*ddt.Node{n0}, false)
	defer teardownSearchTest(tmp, s)

	rs, _, _, _, _ := s.FilterSearch("needs")
	expectFilterSearchResult(t, rs, "Color Definition")

	rs, _, _, _, _ = s.FilterSearch("images")
	expectFilterSearchResult(t, rs, "Color Definition")

	rs, _, _, _, _ = s.FilterSearch("needs images")
	expectFilterSearchResult(t, rs, "Color Definition")
}

// Search test helpers:

func newTestNode(path string, root string) *ddt.Node {
	return ddt.NewNode(path, root, config.NewStaticDB("example"), meta.NewNoopDB(), author.NewNoopDB())
}

func setupSearchTest(t *testing.T, tmp string, lang string, nodes []*ddt.Node, dumpIndex bool) *Search {
	t.Helper()

	var wideIndex bleve.Index
	var narrowIndex bleve.Index

	if dumpIndex {
		searchPath, _ := ioutil.TempDir("", "dsk"+t.Name())
		wideIndex, narrowIndex, _ = NewIndexes(searchPath, lang, true)
	} else {
		wideIndex, narrowIndex, _ = NewIndexes("", lang, false)
	}

	lookup := make(map[string]*ddt.Node)
	for _, n := range nodes {
		lookup[n.URL()] = n
	}

	s := &Search{
		getNode: func(url string) (bool, *ddt.Node, error) {
			if n, ok := lookup[url]; ok {
				return true, n, nil
			}
			return false, nil, nil
		},
		getAllNodes: func() []*ddt.Node {
			ns := make([]*ddt.Node, 0)

			for _, n := range lookup {
				ns = append(ns, n)
			}
			return ns
		},
		getTreeHash: func() (string, error) {
			return "<node-tree-hash>", nil
		},
		lang:        lang,
		wideIndex:   wideIndex,
		narrowIndex: narrowIndex,
	}
	s.IndexTree()
	return s
}

func teardownSearchTest(tmp string, s *Search) {
	s.Close()
	os.RemoveAll(tmp)
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

func expectFilterSearchResult(t *testing.T, nodes []*ddt.Node, url string) {
	t.Helper()

	for _, node := range nodes {
		if node.URL() == url {
			t.Logf("Found node URL '%s' in results", url)
			return
		}
	}
	t.Errorf("Expected '%s', but not included in results", url)
}

func expectNoFilterSearchResult(t *testing.T, nodes []*ddt.Node, url string) {
	t.Helper()

	for _, n := range nodes {
		if n.URL() == url {
			t.Errorf("Not expected '%s' to be included in results", url)
		}
	}
}
