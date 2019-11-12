// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

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
			return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get, "test")

	expected := map[string]string{
		"<a href=\"foo/bar\"></a>":  "<a href=\"/tree/foo/bar?v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"/foo/bar\"></a>": "<a href=\"/tree/foo/bar?v=test\" data-node=\"foo/bar\"></a>",
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
		if url == "foo/bar/baz" || url == "foo/bar" {
			return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get, "test")

	expected := map[string]string{
		"<a href=\"\"></a>":           "<a href=\"/tree/foo/bar?v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\".\"></a>":          "<a href=\"/tree/foo/bar?v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"./\"></a>":         "<a href=\"/tree/foo/bar?v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"baz\"></a>":        "<a href=\"/tree/foo/bar/baz?v=test\" data-node=\"foo/bar/baz\"></a>",
		"<a href=\"./baz\"></a>":      "<a href=\"/tree/foo/bar/baz?v=test\" data-node=\"foo/bar/baz\"></a>",
		"<a href=\"../bar/baz\"></a>": "<a href=\"/tree/foo/bar/baz?v=test\" data-node=\"foo/bar/baz\"></a>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}

func TestTransformAssets(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "asset0.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	get := func(url string) (bool, *Node, error) {
		if url == "foo" {
			return true, &Node{root: tmp, Path: filepath.Join(tmp, url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo", get, "test")

	expected := map[string]string{
		"<a href=\"/foo/asset0.json\">":                  "<a href=\"/tree/foo/asset0.json?v=test\" data-node=\"foo\" data-node-asset=\"asset0.json\">",
		"<a href=\"foo/asset0.json\">":                   "<a href=\"/tree/foo/asset0.json?v=test\" data-node=\"foo\" data-node-asset=\"asset0.json\">",
		"<img src=\"/foo/asset0.json\">":                 "<img src=\"/tree/foo/asset0.json?v=test\" data-node=\"foo\" data-node-asset=\"asset0.json\">",
		"<img src=\"foo/asset0.json\">":                  "<img src=\"/tree/foo/asset0.json?v=test\" data-node=\"foo\" data-node-asset=\"asset0.json\">",
		"<figure><img src=\"foo/asset0.json\"></figure>": "<figure><img src=\"/tree/foo/asset0.json?v=test\" data-node=\"foo\" data-node-asset=\"asset0.json\"></figure>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}

func TestTransformAssetsInComponentAttributes(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	asset0 := filepath.Join(node0, "colors.json")
	ioutil.WriteFile(asset0, []byte(""), 0666)

	get := func(url string) (bool, *Node, error) {
		if url == "foo" {
			return true, &Node{root: tmp, Path: filepath.Join(tmp, url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo", get, "test")

	expected := map[string]string{
		"<ColorGroup src=\"colors.json\"></ColorGroup>": "<colorgroup src=\"/tree/foo/colors.json?v=test\" data-node=\"foo\" data-node-asset=\"colors.json\"></ColorGroup>",
		"<ColorSpecimen src=\"colors.json\" />":         "<colorspecimen src=\"/tree/foo/colors.json?v=test\" data-node=\"foo\" data-node-asset=\"colors.json\"/>",
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
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get, "test")

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

func TestNodeLinksKeepFragmentAndQueryFromOriginal(t *testing.T) {
	get := func(url string) (bool, *Node, error) {
		if url == "foo/bar" {
			return true, &Node{root: "/tmp/xyz", Path: filepath.Join("/tmp/xyz", url)}, nil
		}
		return false, &Node{}, nil
	}
	dt, _ := NewNodeDocTransformer("/tree", "foo/bar", get, "test")

	expected := map[string]string{
		"<a href=\"?answer=42\"></a>":      		"<a href=\"/tree/foo/bar?answer=42&v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"./?answer=42\"></a>":      		"<a href=\"/tree/foo/bar?answer=42&v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"/foo/bar?answer=42\"></a>":      "<a href=\"/tree/foo/bar?answer=42&v=test\" data-node=\"foo/bar\"></a>",
		"<a href=\"/foo/bar#life\"></a>":           "<a href=\"/tree/foo/bar?v=test#life\" data-node=\"foo/bar\"></a>",
		"<a href=\"/foo/bar#life?answer=42\"></a>": "<a href=\"/tree/foo/bar?v=test#life?answer=42\" data-node=\"foo/bar\"></a>",
	}
	for h, e := range expected {
		r, _ := dt.ProcessHTML([]byte(h))

		if !reflect.DeepEqual(r, []byte(e)) {
			t.Errorf("\nexpected input : %s\nto parse to    : %s\nbut got instead: %s", h, e, r)
		}
	}
}
