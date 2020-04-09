// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTitleDerivation(t *testing.T) {
	expected := map[string]string{
		"/tmp/xyz/foo":      "foo",
		"/tmp/xyz/1_foo":    "foo",
		"/tmp/xyz/1-foo":    "foo",
		"/tmp/xyz/0001-foo": "foo",
		"/tmp/xyz/Foo":      "Foo",
	}
	for path, e := range expected {
		n := &Node{Path: path}
		r := n.Title()
		if e != r {
			t.Errorf("\nexpected: %s, result: %s", e, r)
		}
	}
}

func TestCleanURLs(t *testing.T) {
	expected := map[string]string{
		"/bar/xyz/foo":      "xyz/foo",
		"/bar/xyz/1_foo":    "xyz/foo",
		"/bar/xyz/1-foo":    "xyz/foo",
		"/bar/xyz/0001-foo": "xyz/foo",
		"/bar/xyz/Foo":      "xyz/Foo",
		"/bar/02_xyz/1_foo": "xyz/foo",
	}
	for path, e := range expected {
		n := &Node{Path: path, root: "/bar"}
		r := n.URL()
		if e != r {
			t.Errorf("\nexpected: %s, result: %s", e, r)
		}
	}
}

func TestCrumbURLs(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", Path: "/tmp/xyz/foo/bar/baz/"}
	result := n.Crumbs(get)

	expected := []string{
		"foo",
		"foo/bar",
		"foo/bar/baz",
	}

	if len(result) != len(expected) {
		t.Errorf("crumbs do not have the expected length: %d", len(expected))
	}
	for k, v := range expected {
		if result[k].URL() != v {
			t.Errorf("failed to parse crumbs, expectation for key %d failed", k)
			t.Logf("expected: %s, result: %s", v, result[k].URL())
		}
	}
}

func TestCrumbSimpleTitles(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", Path: "/tmp/xyz/foo/bar/baz/"}
	result := n.Crumbs(get)

	expected := []string{
		"foo",
		"bar",
		"baz",
	}

	if len(result) != len(expected) {
		t.Errorf("crumbs do not have the expected length: %d", len(expected))
	}
	for k, v := range expected {
		if result[k].Title() != v {
			t.Errorf("failed to parse crumbs, expectation for key %d failed", k)
			t.Logf("expected: %s, result: %s", v, result[k].URL())
		}
	}
}

func TestCrumbOrderedTitles(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", Path: "/tmp/xyz/01_foo/2-bar/baz/"}
	result := n.Crumbs(get)

	expected := []string{
		"foo",
		"bar",
		"baz",
	}

	if len(result) != len(expected) {
		t.Errorf("crumbs do not have the expected length: %d", len(expected))
	}
	for k, v := range expected {
		if result[k].Title() != v {
			t.Errorf("failed to parse crumbs, expectation for key %d failed", k)
			t.Logf("expected: %s, result: %s", v, result[k].URL())
		}
	}
}

func TestTitlesWithDecomposedFilenames(t *testing.T) {
	n := &Node{Path: "/bar/Cafe\u0301", root: "/bar"}
	if n.Title() != "Café" {
		t.Errorf("failed to decode folder name, got %v", n.Title())
	}
}

func TestAsset(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "asset.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	n := &Node{Path: node0}

	ok, _, err := n.Asset("asset.json")
	if !ok {
		t.Errorf("failed to find asset: %s", err)
	}
}

func TestAssetWithOrderNumberPrefixNoopWithPrefix(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "02_asset.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	n := &Node{Path: node0}

	ok, _, err := n.Asset("02_asset.json")
	if !ok {
		t.Errorf("failed to find asset: %s", err)
	}

	ok, a, _ := n.Asset("asset.json")
	if ok {
		t.Errorf("found asset, where it should not: %s", a)
	}

	ok, a, _ = n.Asset("04_asset.json")
	if ok {
		t.Errorf("found asset, where it should not: %s", a)
	}
}

func TestAssetWithOrderNumberPrefixNoopCollisions(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "02_asset.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	asset1 := filepath.Join(node0, "04_asset.json")
	ioutil.WriteFile(asset1, []byte(""), 0666)

	n := &Node{Path: node0}

	_, a, err := n.Asset("04_asset.json")
	if err != nil {
		t.Error(err)
	}
	if a.Name() != "04_asset.json" {
		t.Errorf("Found wrong asset: %s", a)
	}

	_, a, err = n.Asset("02_asset.json")
	if err != nil {
		t.Error(err)
	}
	if a.Name() != "02_asset.json" {
		t.Errorf("Found wrong asset: %s", a)
	}
}

func TestAssetWithOrderNumberPrefixNoopWithoutPrefix(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "asset.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	n := &Node{Path: node0}

	ok, _, err := n.Asset("asset.json")
	if !ok {
		t.Errorf("failed to find asset: %s", err)
	}

	// Not testing for side-effects that not desired but are okay for
	// us, i.e requesting 02_asset.json and having a match.
}

func TestAssetWithDecomposedFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "Cafe\u0301.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	n := &Node{Path: node0}

	ok, _, err := n.Asset("Cafe\u0301.json")
	if !ok {
		t.Errorf("failed to find asset: %s", err)
	}

	ok, _, err = n.Asset("Café.json")
	if !ok {
		t.Errorf("failed to find asset: %s", err)
	}
}
func BenchmarkHashCalculation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Do not use constructor, so we don't also measure meta parsing.
		n := &Node{
			root:     "example",
			Path:     "example",
			Children: make([]*Node, 0),
		}
		n.CalculateHash()
	}
}
