// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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
