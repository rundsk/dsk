// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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
		n := &Node{path: path}
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
		n := &Node{path: path, root: "/bar"}
		r := n.URL()
		if e != r {
			t.Errorf("\nexpected: %s, result: %s", e, r)
		}
	}
}

func TestCrumbURLs(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		return true, &Node{root: "/tmp/xyz", path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", path: "/tmp/xyz/foo/bar/baz/"}
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
		return true, &Node{root: "/tmp/xyz", path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", path: "/tmp/xyz/foo/bar/baz/"}
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
		return true, &Node{root: "/tmp/xyz", path: filepath.Join("/tmp/xyz", url)}, nil
	}

	n := &Node{root: "/tmp/xyz", path: "/tmp/xyz/01_foo/2-bar/baz/"}
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

func BenchmarkHashCalculation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Do not use constructor, so we don't also measure meta parsing.
		n := &Node{
			root:     "example",
			path:     "example",
			Children: make([]*Node, 0),
		}
		n.Hash()
	}
}

func TestTitlesWithDecomposedFilenames(t *testing.T) {
  n := &Node{path: "/bar/Cafe\u0301", root: "/bar"}
  if n.Title() != "Café" {
    t.Errorf("failed to decode folder name, got %v", n.Title())
  }
}
