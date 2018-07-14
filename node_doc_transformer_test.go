// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNodeLinkAbsolute(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		if url == "foo/bar" {
			return true, &Node{root: "/tmp/xyz", path: filepath.Join("/tmp/xyz", url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get)

	expected := map[string]string{
		"<a href=\"foo/bar\"></a>":  "<a href=\"/tree/foo/bar\" data-node=\"foo/bar\"></a>",
		"<a href=\"/foo/bar\"></a>": "<a href=\"/tree/foo/bar\" data-node=\"foo/bar\"></a>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}

func TestTransformNodeLinkRelative(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		if url == "foo/bar/baz" {
			return true, &Node{root: "/tmp/xyz", path: filepath.Join("/tmp/xyz", url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get)

	expected := map[string]string{
		"<a href=\"baz\"></a>":        "<a href=\"/tree/foo/bar/baz\" data-node=\"foo/bar/baz\"></a>",
		"<a href=\"./baz\"></a>":      "<a href=\"/tree/foo/bar/baz\" data-node=\"foo/bar/baz\"></a>",
		"<a href=\"../bar/baz\"></a>": "<a href=\"/tree/foo/bar/baz\" data-node=\"foo/bar/baz\"></a>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}

func TestTransformNodeAssets(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "asset0.txt")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	get := func(url string) (bool, *Node, error) {
		if url == "foo" {
			return true, &Node{root: tmp, path: filepath.Join(tmp, url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo", get)

	expected := map[string]string{
		"<a href=\"/foo/asset0.txt\">":  "<a href=\"/tree/foo/asset0.txt\" data-node=\"foo\" data-node-asset=\"asset0.txt\">",
		"<a href=\"foo/asset0.txt\">":   "<a href=\"/tree/foo/asset0.txt\" data-node=\"foo\" data-node-asset=\"asset0.txt\">",
		"<img src=\"/foo/asset0.txt\">": "<img src=\"/tree/foo/asset0.txt\" data-node=\"foo\" data-node-asset=\"asset0.txt\">",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}

func TestNonNodeLinks(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get)

	expected := map[string]string{
		"<a href=\"https://example.org\"></a>": "<a href=\"https://example.org\"></a>",
		"<a href=\"/some/other/page\"></a>":    "<a href=\"/some/other/page\"></a>",
		"<a href=\"some/other/page\"></a>":     "<a href=\"/tree/foo/bar/some/other/page\"></a>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}
