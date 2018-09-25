// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func NewRepository(mainPath string, subPath string) (*Repository, error) {
	var path string
	var repo *git.Repository

	path = mainPath
	repo, err := git.PlainOpen(mainPath)
	if err != nil {
		return nil, err
	}

	if subPath != mainPath {
		wt, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		subs, err := wt.Submodules()
		if err != nil {
			return nil, err
		}
		for _, sub := range subs {
			if filepath.Join(mainPath, sub.Config().Path) == subPath {
				subRepo, err := sub.Repository()
				if err != nil {
					return nil, err
				}
				path = subPath
				repo = subRepo
			}
		}
	}
	return &Repository{
		Repository: repo,
		path:       path,
		lookup:     make(map[string]time.Time, 0),
		ticker:     time.NewTicker(5 * time.Second),
		done:       make(chan bool),
	}, nil
}

type Repository struct {
	sync.RWMutex
	*git.Repository

	// Lookup table, mapping file paths to modified times.
	lookup map[string]time.Time

	// Current head reference.
	head *plumbing.Reference

	// Root of the repository's worktree.
	path string

	// Ticker which triggers a lookup rebuild.
	ticker *time.Ticker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

func (r *Repository) StartLookupBuilder() {
	yellow := color.New(color.FgYellow)

	go func() {
		for {
			select {
			case <-r.ticker.C:
				if r.IsLookupStale() {
					if err := r.BuildLookup(); err != nil {
						log.Print(yellow.Sprintf("Failed to rebuild repository lookup table: %s", err))
						continue
					}
				}
			case <-r.done:
				log.Print("Stopping repo lookup builder (received quit)...")
				return
			}
		}
	}()
}

func (r *Repository) StopLookupBuilder() {
	r.done <- true
}

func (r *Repository) Close() {
	r.ticker.Stop()
}

func (r *Repository) IsLookupStale() bool {
	r.RLock()
	defer r.RUnlock()

	if r.head == nil {
		return false
	}
	ref, _ := r.Head()
	return r.head.Hash() != ref.Hash()
}

// BuildLookup will build the lookup table. This allows lookups of
// a file's modified time. Will add modified time for all files and
// directories discovered in root, which is recursively walked.
//
// Implementation based upon snippet provided in:
// https://github.com/src-d/go-git/issues/604
//
// Also see:
// https://github.com/src-d/go-git/issues/417
// https://github.com/src-d/go-git/issues/826
func (r *Repository) BuildLookup() error {
	r.Lock()
	defer r.Unlock()

	start := time.Now()
	pathsCached := make(map[string]bool, 0)

	ref, _ := r.Head()
	if ref == nil {
		log.Printf("No commits in repository %s, yet", r.path)
		return nil
	}
	r.head = ref

	err := filepath.Walk(r.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(r.path) == f.Name()

			if IgnoreNodesRegexp.MatchString(f.Name()) && !isRoot {
				return filepath.SkipDir
			}
			return nil // Git only knows about files
		}
		rel, _ := filepath.Rel(r.path, path)
		pathsCached[rel] = false
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", r.path, err)
	}

	r.lookup = make(map[string]time.Time, 0)

	commits, err := r.Log(&git.LogOptions{From: r.head.Hash()})
	if err != nil {
		return err
	}
	defer commits.Close()

	var prevCommit *object.Commit
	var prevTree *object.Tree

Outer:
	for {
		commit, err := commits.Next()
		if err != nil {
			break
		}
		currentTree, err := commit.Tree()
		if err != nil {
			return err
		}

		if prevCommit == nil {
			prevCommit = commit
			prevTree = currentTree
			continue
		}

		changes, err := currentTree.Diff(prevTree)
		if err != nil {
			return err
		}

		for _, c := range changes {
			if c.To.Name == "" {
				continue
			}
			if isCached, ok := pathsCached[c.To.Name]; !ok || isCached {
				// Not interested in this file.
				continue
			}
			r.lookup[c.To.Name] = prevCommit.Author.When
			pathsCached[c.To.Name] = true

			if len(r.lookup) >= len(pathsCached) {
				break Outer
			}
		}

		prevCommit = commit
		prevTree = currentTree
	}

	log.Printf("Built repository lookup table in %s with %d entries", time.Since(start), len(r.lookup))
	return nil
}

// Modified considers any changes inside given path or its
// subdirectories as a change to the path.
func (r *Repository) Modified(path string) (time.Time, error) {
	// Fast path for files.
	if m, ok := r.lookup[path]; ok {
		return m, nil
	}

	var modified time.Time

	for p, m := range r.lookup {
		if !filepath.HasPrefix(p, path) {
			continue
		}
		if m.After(modified) {
			modified = m
		}
	}

	if !modified.IsZero() {
		return modified, nil
	}
	if r.head == nil {
		return modified, nil
	}
	// When there's only one commit no diffing has been taken place.
	// It can be assumed that this is an initial commit adding all
	// files.
	commit, err := r.CommitObject(r.head.Hash())
	if err != nil {
		return modified, err
	}
	return commit.Author.When, nil
}
