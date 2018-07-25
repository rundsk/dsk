// Copyright 2017 Atelier Disko. All rights reserved.
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

func TestDocTitlesWithDecomposedFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "01-Cafe\u0301.md")
	ioutil.WriteFile(doc0, []byte(""), 0666)

	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}

 	docs, err := node.Docs()
	if err != nil {
		t.Errorf("can’t read docs")
	}

	if docs[0].Title() != "Café" {
    t.Errorf("failed to decode file name, got %v", docs[0].Title())
  }
}
