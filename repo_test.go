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
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestRepoLastModified(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "repo")
	defer os.RemoveAll(tmp)

	gr, _ := git.PlainInit(tmp, false)
	w, _ := gr.Worktree()

	path := filepath.Join(tmp, "Diversity")
	n := NewNode(path, tmp)
	n.Create()
	n.CreateDoc("doc0.md", []byte("a"))
	n.CreateDoc("doc1.md", []byte("a"))

	repo, err := NewRepository(tmp, "")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(1981, 8, 11, 12, 0, 0, 0, time.UTC)

	w.Add("Diversity/doc0.md")
	w.Commit("message0", &git.CommitOptions{
		Author: &object.Signature{
			When: now,
		},
	})
	repo.BuildCache()

	result, err := repo.Modified("Diversity")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Equal(now) {
		t.Errorf("%s != %s", result, now)
	}

	now = now.Add(time.Hour * 2)

	w.Add("Diversity/doc1.md")
	w.Commit("message1", &git.CommitOptions{
		Author: &object.Signature{
			When: now,
		},
	})
	repo.BuildCache()

	result, err = repo.Modified("Diversity")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Equal(now) {
		t.Errorf("%s != %s", result, now)
	}
}
