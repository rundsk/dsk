// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rundsk/dsk/internal/bus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// NewRepo initializes a new Repo. A mainpath must always be given,
// an optional subpath may be given when submodules are in use.
// Optionally an existing Repository may be provided, if none is
// provided one will be created on the fly.
func NewRepo(mainpath string, subpath string, gr *git.Repository, b *bus.Broker) (*Repo, error) {
	log.Printf("Initializing repo on %s...", mainpath)

	var path string
	var repo *git.Repository

	path = mainpath

	if gr == nil {
		gr, err := git.PlainOpen(path)
		if err != nil {
			return nil, err
		}
		repo = gr
	} else {
		repo = gr
	}

	// When subpath is provided it points to the root of a submodule
	// inside the main repository. We can only access the Repository
	// object indirectly by iterating over the list of submodules as
	// retrieved from the main repository.
	//
	// Not being able to match up subpath of a submodule with one of
	// the paths listed by the main repository is an error condition,
	// it means something bad happened while auto-detecting the roots
	// of the repository and/or submodule.
	//
	// Please note that a submodule must not directly contain the DDT
	// so we cannot use the subpath as an interest filter for building
	// lookup tables.
	if subpath != "" && subpath != mainpath {
		var hasFoundMatchingSub bool

		wt, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		subs, err := wt.Submodules()
		if err != nil {
			return nil, err
		}
		if len(subs) == 0 {
			return nil, errors.New("no submodules available. Are you missing a .gitmodules file?")
		}
		for _, sub := range subs {
			if filepath.Join(mainpath, sub.Config().Path) != subpath {
				log.Printf("Skipping submodule at %s", filepath.Join(mainpath, sub.Config().Path))
				continue
			}
			subrepo, err := sub.Repository()
			if err != nil {
				return nil, err
			}
			path = subpath
			repo = subrepo

			hasFoundMatchingSub = true
		}
		if !hasFoundMatchingSub {
			return nil, fmt.Errorf("failed to match subrepo %s to available ones", subpath)
		}
	}

	r := &Repo{
		repo:   repo,
		Path:   path,
		broker: b,
		ticker: time.NewTicker(2 * time.Second),
		done:   make(chan bool),
	}

	r.versionsLookup, _ = NewLookup(fmt.Sprintf("%s versions", r), func() (*plumbing.Reference, interface{}, error) {
		return r.buildVersionsLookup()
	})
	r.modifiedLookup, _ = NewLookup(fmt.Sprintf("%s modified", r), func() (*plumbing.Reference, interface{}, error) {
		return r.buildModifiedLookup()
	})

	ref, err := r.repo.Head()
	if ref != nil && err != nil {
		return r, err
	}
	r.head = ref

	// We must determine the current reference name now, as once the
	// head moves we cannot match tag references up anymore.
	if r.head != nil {
		currefn, err := r.currentReferenceName()
		if err != nil {
			return r, err
		}
		r.name = currefn
	}

	return r, r.Open()
}

type Repo struct {
	sync.RWMutex

	// repo is the object we wrap. It is protected by a mutex, as the
	// underlying library is not thread safe by design. We can't use a
	// RWMutex as we don't know if an action causes an internal write
	// or just a read, i.e. what looks from the outside like a read
	// action may in fact be an internal cache write.
	repo      *git.Repository
	repoMutex sync.Mutex

	// Root of the repository's worktree.
	Path string

	// name is the current reference name, i.e. refs/tags/v1.2.3.
	name plumbing.ReferenceName

	// head is the last known head reference to this struct, it is
	// used to check if the head moved and the repository changed.
	head *plumbing.Reference

	// Ticker which eventually triggers a lookup rebuild.
	ticker *time.Ticker

	// broker is an external broker where we send events to, i.e. on repo.changed or repo.ready.
	broker *bus.Broker

	versionsLookup *Lookup

	modifiedLookup *Lookup

	// done is a quit channel, receiving true, when we are closed.
	done chan bool
}

func (r *Repo) String() string {
	return fmt.Sprintf("repo (...%s)", filepath.Base(r.Path))
}

func (r *Repo) Open() error {
	go func() {
		for {
			select {
			case <-r.ticker.C:
				if r.HasHeadChanged() {
					r.repoMutex.Lock()
					ref, _ := r.repo.Head()
					r.repoMutex.Unlock()

					r.broker.Accept("repo.changed", ref.Name().Short())

					// Ensure we detect head change only once, the
					// stale detection is not influenced by this.
					r.Lock()
					r.head = ref
					r.Unlock()

					r.versionsLookup.RequestBuild(ref)
					r.modifiedLookup.RequestBuild(ref)
				}
			case <-r.done:
				log.Printf("Stopping %s periodic change detector (received quit)...", r)
				return
			}
		}
	}()

	return nil
}

func (r *Repo) Close() error {
	r.ticker.Stop()
	r.done <- true

	r.versionsLookup.Close()
	r.modifiedLookup.Close()

	return nil
}

func (r *Repo) HasHeadChanged() bool {
	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	if ref == nil {
		return false
	}
	return !r.isSameRef(ref, r.head)
}

func (r *Repo) IsStale() bool {
	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	if ref == nil {
		return false
	}
	return r.modifiedLookup.IsStale(ref) || r.versionsLookup.IsStale(ref)
}

func (r *Repo) isSameRef(a *plumbing.Reference, b *plumbing.Reference) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Hash() == b.Hash()
}

func (r *Repo) currentReferenceName() (plumbing.ReferenceName, error) {
	var found plumbing.ReferenceName

	r.repoMutex.Lock()
	defer r.repoMutex.Unlock()

	head, err := r.repo.Head()
	if err != nil {
		return found, err
	}
	refs, err := r.repo.References()
	if err != nil {
		return found, err
	}
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			if ref.Hash() == head.Hash() {
				found = ref.Name()
				return storer.ErrStop
			}
		}
		return nil
	})
	return found, err
}

func (r *Repo) HasUpstreamChanged() (bool, error) {
	r.repoMutex.Lock()
	defer r.repoMutex.Unlock()

	err := r.repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateFromUpstream updates the currently checked out
// reference from upstream.
//
// TODO: This currently fails with "object not found" or "already
// up-to-date", or "not-fast-forward", when this shouldn't be the
// case. We probably have to switch to fetch+hard reset.
func (r *Repo) UpdateFromUpstream() error {
	log.Printf("Updating %s from upstream...", r)
	start := time.Now()

	r.repoMutex.Lock()
	defer r.repoMutex.Unlock()

	ref, _ := r.repo.Head()

	r.RLock()
	rname := r.name
	r.RUnlock()

	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}
	err = wt.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: rname,
		SingleBranch:  true,
		// When we remove the HEAD commit from the live repo, we want
		// its versions to be reset.
		Force: true,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Printf("%s is already up to date with upstream", r)
			return nil
		}
		return fmt.Errorf("update of %s from upstream failed: %s", r, err)
	}

	r.Lock()
	r.head = ref
	r.Unlock()

	log.Printf("Updated %s from upstream in %s", r, time.Since(start))
	return nil
}

// Modified considers any changes in and below given path as a change
// to the path. The path must be absolute and rooted at the repository
// path.
func (r *Repo) ModifiedWithContext(ctx context.Context, path string) (time.Time, error) {
	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	var lookup map[string]time.Time
	var modified time.Time

	select {
	case res := <-r.modifiedLookup.GetDirtyOkay(ref):
		lookup = res.(map[string]time.Time)
	case <-ctx.Done():
		return modified, errors.New("no data")
	}

	path, err := filepath.Rel(r.Path, path)
	if err != nil {
		return modified, err
	}

	// Fast path for files.
	if m, ok := lookup[path]; ok {
		return m, nil
	}

	for p, m := range lookup {
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
	// When there's only one commit no diffing has been taken place.
	// It can be assumed that this is an initial commit adding all
	// files.
	r.repoMutex.Lock()
	commit, err := r.repo.CommitObject(ref.Hash())
	r.repoMutex.Unlock()
	if err != nil {
		return modified, err
	}
	return commit.Author.When, nil
}

// Version returns the current checked out version.
func (r *Repo) Version() (*Version, error) {
	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	return NewVersionFromRef(ref), nil
}

// Versions are sorted tags (we remove a leading "v") and branches
// (which are prefixed by "dev-", as to avoid collisions with tag
// names).
func (r *Repo) Versions() (*Versions, error) {
	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	return (<-r.versionsLookup.GetDirtyOkay(ref)).(*Versions), nil
}

// BuildModifiedLookup will build and swap out the lookup table for
// looking up modified times on a file. Will add modified time for
// all files and directories discovered in root, which is recursively
// walked.
//
// Implementation based upon snippet provided in:
// https://github.com/src-d/go-git/issues/604
//
// Also see:
// https://github.com/src-d/go-git/issues/417
// https://github.com/src-d/go-git/issues/826
func (r *Repo) buildModifiedLookup() (*plumbing.Reference, map[string]time.Time, error) {
	pathsCached := make(map[string]bool, 0)
	lookup := make(map[string]time.Time, 0)

	r.repoMutex.Lock()
	ref, _ := r.repo.Head()
	r.repoMutex.Unlock()

	if ref == nil {
		log.Printf("No commits in repo %s, yet", r.Path)
		return ref, lookup, nil
	}

	err := filepath.Walk(r.Path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			isRoot := filepath.Base(r.Path) == f.Name()

			if strings.HasPrefix(f.Name(), ".") && !isRoot {
				return filepath.SkipDir
			}
			return nil // Git only knows about files
		}
		rel, _ := filepath.Rel(r.Path, path)
		pathsCached[rel] = false
		return nil
	})
	if err != nil {
		return ref, lookup, fmt.Errorf("failed to walk directory tree %s: %s", r.Path, err)
	}

	r.repoMutex.Lock()
	defer r.repoMutex.Unlock()

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

func (r *Repo) buildVersionsLookup() (*plumbing.Reference, *Versions, error) {
	r.repoMutex.Lock()
	defer r.repoMutex.Unlock()

	ref, _ := r.repo.Head()
	versions := &Versions{}

	iter, err := r.repo.References()
	if err != nil {
		return ref, versions, err
	}

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsTag() {
			versions.Add(NewVersionFromRef(ref))
		}
		return nil
	})
	return ref, versions, err
}
