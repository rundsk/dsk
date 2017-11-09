// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"
)

func TestCrumbs(t *testing.T) {
	n := &Node{URL: "foo/bar/baz"}

	result := n.CrumbURLs()
	expected := []string{
		"foo",
		"foo/bar",
		"foo/bar/baz",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Logf("%+v != %+v", result, expected)
		t.Error("failed to parse crumbs")
	}
}
