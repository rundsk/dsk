// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plex

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/rundsk/dsk/internal/author"
	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/config"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/meta"
	"github.com/rundsk/dsk/internal/notify"
	"github.com/rundsk/dsk/internal/search"
	"github.com/rundsk/dsk/internal/vcs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

type sourceCompleteFunc func(*Source) (string, *git.Repository, error)

// NewSource initializes a new source and Open()s it to ready it.
func NewSource(name string, path string, c config.DB) (*Source, error) {
	s := &Source{
		Name:     name,
		Path:     path,
		ConfigDB: c,
	}
	s.Teardown = &Teardown{Scope: s.String()}

	return s, s.Open(nil)
}

func NewLazySource(name string, completeFn sourceCompleteFunc, c config.DB) (*Source, error) {
	s := &Source{
		Teardown:   &Teardown{Scope: fmt.Sprintf("%s source", name)},
		Name:       name,
		completeFn: completeFn,
		ConfigDB:   c,
	}

	b, err := bus.NewBroker()
	if err != nil {
		return s, err
	}
	s.Broker = b
	s.Teardown.AddFunc(b.Close)

	return s, nil
}

type Source struct {
	*Teardown

	// Name identifies the source, usually this is the name of the
	// version the source belongs to.
	Name string

	// Path is the absolute path to the DDT, this might be a subpath
	// of Repo.Path, when the DDT is nested inside a repository.
	// Optional when a completeFn is provided.
	Path string

	// completeFn completes the source, the source initizalitation can
	// then be finalized using Open().
	completeFn sourceCompleteFunc

	// Broker is the Source-specific broker.
	Broker *bus.Broker

	// ConfigDB is the central configuration and managed from the outside.
	ConfigDB config.DB

	Tree *ddt.Tree

	Search *search.Search

	MetaDB meta.DB

	AuthorDB author.DB

	// Repo, optional.
	Repo *vcs.Repo

	// Watcher watches the Path for changes, optional.
	Watcher *notify.Watcher
}

func (s *Source) String() string {
	if s.IsComplete() {
		return fmt.Sprintf("source (%s)", s.Name)
	}
	return fmt.Sprintf("incomplete source (%s)", s.Name)
}

func (s *Source) IsComplete() bool {
	return s.Path != ""
}

func (s *Source) Complete() error {
	log.Printf("Completing %s...", s)
	start := time.Now()
	defer s.Broker.Accept("source.status.changed", "completed")

	p, r, err := s.completeFn(s)
	if err != nil {
		return err
	}
	s.Path = p
	err = s.Open(r)

	log.Printf("Completed %s in %s", s, time.Since(start))
	return err
}

// Open continues the initialization of the struct. When an optinal
// existing git.Repository is provided it will be used instead of
// initialization a new one using path information of the struct.
func (s *Source) Open(gr *git.Repository) error {
	if s.Broker == nil {
		b, err := bus.NewBroker()
		if err != nil {
			return err
		}
		s.Broker = b
		s.Teardown.AddFunc(b.Close)
	}

	if gr != nil {
		r, err := vcs.NewRepo(
			s.Path,
			"",
			gr,
			s.Broker,
		)
		if err != nil {
			return err
		}
		s.Repo = r
		s.Teardown.AddFunc(s.Repo.Close)

		log.Printf("Synthesizing fs.changed events from repo.changed for %s...", s)
		s.Broker.SubscribeFunc("repo.changed", func() error {
			s.Broker.Accept("fs.changed", "repo.changed")
			return nil
		})
	} else {
		ok, rroot, rsub, err := vcs.FindRepo(s.Path)
		if err != nil {
			return err
		}
		if ok {
			r, err := vcs.NewRepo(
				rroot,
				rsub,
				nil,
				s.Broker,
			)
			if err != nil {
				return err
			}
			s.Repo = r
			s.Teardown.AddFunc(s.Repo.Close)
		}

		log.Printf("Using filesystem watcher for %s...", s)
		w, err := notify.NewWatcher(s.Path)
		if err != nil {
			return err
		}
		s.Watcher = w
		s.Teardown.AddFunc(s.Watcher.Close)

		done := s.Broker.Connect(s.Watcher.Subscribable, "fs")
		s.Teardown.AddChan(done)
	}

	if s.Repo != nil {
		s.MetaDB = meta.NewChainDB(meta.NewRepoDB(s.Repo), meta.NewFSDB())
	} else {
		log.Printf("Will configure %s without repo", s)
		s.MetaDB = meta.NewFSDB()
	}

	var adb author.DB
	ok, f, err := author.FindFile(s.Path)
	if err != nil {
		return err
	}
	if ok {
		adb, err = author.NewTxtDB(f)
		if err != nil {
			return err
		}
		s.AuthorDB = adb
		s.Teardown.AddFunc(s.AuthorDB.Close)

		done := s.Broker.SubscribeFunc("fs.changed", s.AuthorDB.Refresh)
		s.Teardown.AddChan(done)
	} else {
		s.AuthorDB = author.NewNoopDB()
	}

	t, err := ddt.NewTree(
		s.Path,
		s.ConfigDB,
		s.AuthorDB,
		s.MetaDB,
		s.Broker,
	)
	if err != nil {
		return err
	}
	s.Tree = t

	done := s.Broker.SubscribeFunc("fs.changed", s.Tree.Sync)
	s.Teardown.AddChan(done)

	se, err := search.NewSearch("", t, s.ConfigDB.Data().Lang, false)
	if err != nil {
		return err
	}
	s.Search = se
	s.Teardown.AddFunc(se.Close)

	done = s.Broker.SubscribeFunc("tree.synced", se.Refresh)
	s.Teardown.AddChan(done)

	return nil
}

func (s *Source) Close() error {
	return s.Teardown.Close()
}

// ReverseConnectBroker connects the internal broker to the provided
// one, essentially pushing all internal messages of this source out
// to the external broker.
func (s *Source) ReverseConnectBroker(b *bus.Broker) chan bool {
	return b.Connect(s.Broker.Subscribable, s.Name)
}

// CloneFrom copies the source repository by cloning it into a
// temporary directory.
func (s *Source) CloneFrom(source string, refn plumbing.ReferenceName) (string, *git.Repository, error) {
	log.Printf("Cloning repo from %s...", source)
	start := time.Now()

	tmp, err := ioutil.TempDir("", fmt.Sprintf("dsk%s", s.Name))
	if err != nil {
		return tmp, nil, err
	}
	s.Teardown.AddAsyncFunc(func() error {
		log.Printf("Removing temporary checkout for %s...", s)
		return os.RemoveAll(tmp)
	})

	// We explicitly define how we store the object database and the
	// worktree. We might later want to use a memory-backed object
	// database.
	wt := osfs.New(tmp)

	// odb := memory.NewStorage()
	dot, _ := wt.Chroot(".git")
	odb := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	r, err := git.Clone(odb, wt, &git.CloneOptions{
		URL:           source,
		ReferenceName: refn,
		SingleBranch:  true,
	})
	if err != nil {
		return tmp, r, err
	}
	log.Printf("Cloned repo from %s in %s", source, time.Since(start))
	return tmp, r, err
}

// UpdateFromUpstream updates the source repo from the one it was
// cloned. For the live source, this might be a remote repository.
func (s *Source) UpdateFromUpstream() error {
	if s.Repo == nil {
		return fmt.Errorf("%s has no repo", s)
	}
	return s.Repo.UpdateFromUpstream()
}

func (s *Source) HasMultiVersionsSupport() bool {
	return s.Repo != nil && s.ConfigDB != nil
}

// Versions retrieves versions from Repo.
func (s *Source) Versions() (*vcs.Versions, error) {
	return s.Repo.Versions()
}
