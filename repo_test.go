// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
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
	repo.BuildLookup()

	result, err := repo.Modified(filepath.Join(tmp, "Diversity"))
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
	repo.BuildLookup()

	result, err = repo.Modified(filepath.Join(tmp, "Diversity"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Equal(now) {
		t.Errorf("%s != %s", result, now)
	}
}

func BenchmarkBuildLookup(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	tmp, _ := ioutil.TempDir("", "repo")
	defer os.RemoveAll(tmp)

	gr, _ := git.PlainInit(tmp, false)
	w, _ := gr.Worktree()

	now := time.Date(1981, 8, 11, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 100; i++ {
		path := filepath.Join(tmp, fmt.Sprintf("Diversity%d", i))

		n := NewNode(path, tmp)
		n.Create()
		n.CreateDoc("doc0.md", []byte("a"))
		n.CreateDoc("doc1.md", []byte("a"))

		w.Add(fmt.Sprintf("Diversity%d/doc0.md", i))
		w.Add(fmt.Sprintf("Diversity%d/doc1.md", i))

		w.Commit(fmt.Sprintf("message%d", i), &git.CommitOptions{
			Author: &object.Signature{
				When: now,
			},
		})
	}
	repo, err := NewRepository(tmp, "")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		repo.BuildLookup()
	}
}
