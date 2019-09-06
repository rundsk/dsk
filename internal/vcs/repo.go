// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	ErrNoData       = errors.New("not enough or no data")
	ErrRepoNotFound = errors.New("no repository found")
)

// Searches beginning at given path up, until it finds a directory
// containing a ".git" directory. We differentiate between submodules
// having a ".git" file and regular repositories where ".git" is an
// actual directory.
func FindRepo(treeRoot string, searchSubmodule bool) (string, error) {
	var path = treeRoot

	for path != "." && path != "/" {
		s, err := os.Stat(filepath.Join(path, ".git"))

		if err == nil {
			if searchSubmodule && s.Mode().IsRegular() {
				return path, nil
			} else if !searchSubmodule && s.Mode().IsDir() {
				return path, nil
			}
		}
		path, err = filepath.Abs(path + "/..")
		if err != nil {
			return "", err
		}
	}
	return "", ErrRepoNotFound
}

// NewRepo initializes a new Repo. A mainPath must always
// be given, an optional subPath may be given when submodules are in
// use.
func NewRepo(mainPath string, subPath string, isValidVersion func(string) (bool, error)) (*Repo, error) {
	var path string
	var repo *git.Repository

	path = mainPath
	repo, err := git.PlainOpen(mainPath)
	if err != nil {
		return nil, err
	}

	var hasFoundMatchingSub bool

	if subPath != "" && subPath != mainPath {
		wt, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		subs, err := wt.Submodules()
		if err != nil {
			return nil, err
		}
		if len(subs) == 0 {
			return nil, errors.New("No submodules available. Are you missing a .gitmodules file?")
		}
		for _, sub := range subs {
			if filepath.Join(mainPath, sub.Config().Path) != subPath {
				log.Printf("Skipping submodule at %s", filepath.Join(mainPath, sub.Config().Path))
				continue
			}
			subRepo, err := sub.Repository()
			if err != nil {
				return nil, err
			}
			path = subPath
			repo = subRepo

			hasFoundMatchingSub = true
		}
		if !hasFoundMatchingSub {
			return nil, fmt.Errorf("Failed to match subrepository %s to available ones", subPath)
		}
	}
	return &Repo{
		isValidVersion: isValidVersion,
		repo:           repo,
		path:           path,
		fileMetaLookup: make(map[string]time.Time, 0),
		ticker:         time.NewTicker(5 * time.Second),
		done:           make(chan bool),
	}, nil
}

type Repo struct {
	sync.RWMutex

	isValidVersion func(string) (bool, error)

	repo *git.Repository

	fileMetaLookup map[string]time.Time

	// Current head reference.
	head *plumbing.Reference

	// Root of the repository's worktree.
	path string

	// Ticker which triggers a lookup rebuild.
	ticker *time.Ticker

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

func (r *Repo) StartLookupBuilder() {
	yellow := color.New(color.FgYellow)

	go func() {
		for {
			select {
			case <-r.ticker.C:
				if r.IsLookupStale() {
					if err := r.BuildLookups(); err != nil {
						log.Print(yellow.Sprintf("Failed to rebuild repository lookup tables: %s", err))
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

func (r *Repo) StopLookupBuilder() error {
	r.done <- true
	return nil
}

func (r *Repo) Close() error {
	r.ticker.Stop()
	return nil
}

func (r *Repo) IsLookupStale() bool {
	r.RLock()
	defer r.RUnlock()

	if r.head == nil {
		return true
	}
	ref, _ := r.repo.Head()
	return r.head.Hash() != ref.Hash()
}

func (r *Repo) BuildLookups() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		start := time.Now()

		ref, l, err := r.FileMetaLookup()
		if err != nil {
			log.Printf("Failed to built repository file meta lookup table: %s", err)
			return
		}
		log.Printf("Built repository file meta lookup table with %d object/s in %s", len(l), time.Since(start))

		r.Lock()
		r.head = ref
		r.fileMetaLookup = l
		r.Unlock()

		wg.Done()
	}()

	wg.Wait()
	return nil
}

// ModifiedLookup will build and return the lookup table for looking
// up modified times on a file. Will add modified time for all files
// and directories discovered in root, which is recursively walked.
//
// Implementation based upon snippet provided in:
// https://github.com/src-d/go-git/issues/604
//
// Also see:
// https://github.com/src-d/go-git/issues/417
// https://github.com/src-d/go-git/issues/826
func (r *Repo) FileMetaLookup() (*plumbing.Reference, map[string]time.Time, error) {
	pathsCached := make(map[string]bool, 0)
	lookup := make(map[string]time.Time, 0)
	ref, _ := r.repo.Head()

	if ref == nil {
		log.Printf("No commits in repository %s, yet", r.path)
		return ref, lookup, nil
	}

	err := filepath.Walk(r.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(r.path) == f.Name()

			if strings.HasPrefix(f.Name(), ".") && !isRoot {
				return filepath.SkipDir
			}
			return nil // Git only knows about files
		}
		rel, _ := filepath.Rel(r.path, path)
		pathsCached[rel] = false
		return nil
	})
	if err != nil {
		return ref, lookup, fmt.Errorf("Failed to walk directory tree %s: %s", r.path, err)
	}

	commits, err := r.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return ref, lookup, err
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
			return ref, lookup, err
		}

		if prevCommit == nil {
			prevCommit = commit
			prevTree = currentTree
			continue
		}

		changes, err := currentTree.Diff(prevTree)
		if err != nil {
			return ref, lookup, err
		}

		for _, c := range changes {
			if c.To.Name == "" {
				continue
			}
			if isCached, ok := pathsCached[c.To.Name]; !ok || isCached {
				// Not interested in this file.
				continue
			}
			lookup[c.To.Name] = prevCommit.Author.When
			pathsCached[c.To.Name] = true

			if len(lookup) >= len(pathsCached) {
				break Outer
			}
		}

		prevCommit = commit
		prevTree = currentTree
	}

	return ref, lookup, nil
}

// Modified considers any changes in and below given path as a change
// to the path. The path must be absolute and rooted at the repository
// path.
func (r *Repo) Modified(path string) (time.Time, error) {
	r.RLock()
	defer r.RUnlock()

	var modified time.Time

	path, err := filepath.Rel(r.path, path)
	if err != nil {
		return modified, err
	}

	// Fast path for files.
	if m, ok := r.fileMetaLookup[path]; ok {
		return m, nil
	}

	for p, m := range r.fileMetaLookup {
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
		return modified, ErrNoData
	}
	// When there's only one commit no diffing has been taken place.
	// It can be assumed that this is an initial commit adding all
	// files.
	commit, err := r.repo.CommitObject(r.head.Hash())
	if err != nil {
		return modified, err
	}
	return commit.Author.When, nil
}
