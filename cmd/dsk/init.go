// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/atelierdisko/dsk/internal/author"
	"github.com/atelierdisko/dsk/internal/bus"
	"github.com/atelierdisko/dsk/internal/config"
	"github.com/atelierdisko/dsk/internal/ddt"
	"github.com/atelierdisko/dsk/internal/frontend"
	"github.com/atelierdisko/dsk/internal/notify"
	"github.com/atelierdisko/dsk/internal/pathutil"
	"github.com/atelierdisko/dsk/internal/search"
	"github.com/atelierdisko/dsk/internal/vcs"
	"github.com/fatih/color"
)

func registerNodeTreeSyncedSubscriber(fn func() error, b *bus.Broker) chan bool {
	yellow := color.New(color.FgYellow)

	done := make(chan bool)
	go func() {
		id, messages := b.Subscribe("tree.synced")

		for {
			select {
			case _, ok := <-messages:
				if !ok {
					log.Print("Stopping subscriber (channel closed)...")
					b.Unsubscribe(id)
					return
				}
				if err := fn(); err != nil {
					log.Print(yellow.Sprintf("Failed refreshing: %s", err))
				}
			case <-done:
				log.Print("Stopping subscriber (received quit)...")
				b.Unsubscribe(id)
				return
			}
		}
	}()

	return done
}

func initBroker() (*bus.Broker, CleanupFunc, error) {
	b := bus.NewBroker()

	if err := b.Start(); err != nil {
		return b, b.Close, err
	}

	return b, func() error {
		b.Stop()
		b.Close()
		return nil
	}, nil
}

func initWatcher(here string) (*notify.Watcher, CleanupFunc, error) {
	w := notify.NewWatcher(here)

	if err := w.Start(); err != nil {
		return w, w.Close, err
	}

	return w, func() error {
		w.Stop()
		w.Close()
		return nil
	}, nil
}

func initNodeTree(
	here string,
	cdb *config.DB,
	adb *author.DB,
	r *vcs.Repo,
	w *notify.Watcher,
	b *bus.Broker,
) (*ddt.NodeTree, CleanupFunc, error) {
	t := ddt.NewNodeTree(
		here,
		cdb,
		adb,
		r,
		b,
	)
	done := make(chan bool)

	go func() {
		id, watch := w.Subscribe("fs.changed")

		for {
			select {
			case p := <-watch:
				b.Accept(bus.NewMessage(
					"tree.changed", pathutil.Pretty(p.(string)),
				))
				log.Printf("Syncing node tree...")

				if err := t.Sync(); err != nil {
					log.Printf("Syncing node tree failed: %s", err)
				}
			case <-done:
				log.Print("Stopping node tree syncing (received quit)...")
				w.Unsubscribe(id)
				return
			}
		}
	}()

	return t, func() error {
		done <- true
		return nil
	}, t.Sync()
}

func initConfigDB(here string, b *bus.Broker) (*config.DB, CleanupFunc, error) {
	ok, f, err := config.FindFile(here)
	if err != nil {
		return &config.DB{}, nil, fmt.Errorf("failed to find configuration file: %s", err)
	}
	if !ok {
		return &config.DB{}, nil, nil
	}
	c, err := config.NewDB(f)
	if err != nil {
		return c, c.Close, err
	}
	log.Printf("Loaded configuration from: %s", pathutil.Pretty(f))

	done := registerNodeTreeSyncedSubscriber(c.Refresh, b)

	return c, func() error {
		done <- true
		return c.Close()
	}, nil
}

func initRepo(here string, config *config.DB) (*vcs.Repo, CleanupFunc, error) {
	rroot, err := vcs.FindRepo(here, false)
	if err != nil && err != vcs.ErrRepoNotFound {
		return &vcs.Repo{}, nil, fmt.Errorf("failed to detect repository: %s", err)
	}
	if err == vcs.ErrRepoNotFound {
		return &vcs.Repo{}, nil, nil
	}
	log.Printf("Detected repository support in: %s", rroot)

	rsub, err := vcs.FindRepo(here, true)
	if err == nil {
		log.Printf("Using repository submodule in: %s", rsub)
	}
	if err != nil && err != vcs.ErrRepoNotFound {
		return &vcs.Repo{}, nil, fmt.Errorf("failed to detect repository: %s", err)
	}

	repo, err := vcs.NewRepo(rroot, rsub, config.IsValidVersion)
	if err != nil {
		return repo, repo.Close, fmt.Errorf("failed to enable repository support: %s", err)
	}
	repo.StartLookupBuilder()

	return repo, func() error {
		repo.StopLookupBuilder()
		repo.Close()
		return nil
	}, nil
}

func initAuthorDB(here string, b *bus.Broker) (*author.DB, CleanupFunc, error) {
	ok, f, err := author.FindFile(here)
	if err != nil {
		return &author.DB{}, nil, fmt.Errorf("failed to find authors file: %s", err)
	}
	if !ok {
		return &author.DB{}, nil, nil
	}

	adb, err := author.NewDB(f)
	if err != nil {
		return adb, nil, err
	}
	log.Printf("Loaded authors database from: %s", pathutil.Pretty(f))

	done := registerNodeTreeSyncedSubscriber(adb.Refresh, b)

	return adb, func() error {
		done <- true
		return adb.Close()
	}, nil
}

func initSearch(t *ddt.NodeTree, b *bus.Broker, cdb *config.DB) (*search.Search, CleanupFunc, error) {
	s, err := search.NewSearch("", t, cdb.Data().Lang, false)
	if err != nil {
		return s, nil, err
	}

	done := registerNodeTreeSyncedSubscriber(s.Refresh, b)

	return s, func() error {
		done <- true
		return s.Close()
	}, s.IndexTree()
}

func initFrontend(runtime string, t *ddt.NodeTree) (*frontend.Frontend, CleanupFunc, error) {
	if runtime == "" {
		log.Print("Using built-in frontend")
		return frontend.NewFrontendFromEmbedded(t.Path), nil, nil

	}
	log.Printf("Using runtime frontend from: %s", pathutil.Pretty(runtime))

	fe, err := frontend.NewFrontendFromPath(runtime, t.Path)
	return fe, nil, err
}
