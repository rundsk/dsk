// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plex

import (
	"context"
	"log"
	"path/filepath"

	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/config"
	"github.com/rundsk/dsk/internal/frontend"
	"github.com/rundsk/dsk/internal/notify"
	"github.com/rundsk/dsk/internal/vcs"
	"golang.org/x/text/unicode/norm"
	git "gopkg.in/src-d/go-git.v4"
)

func NewApp(version string, livePath string, componentsPath string, frontendPath string) *App {
	log.Print("Initializing application...")

	return &App{
		Teardown:       &Teardown{Scope: "app"},
		Version:        version,
		livePath:       livePath,
		componentsPath: componentsPath,
		frontendPath:   frontendPath,
	}
}

type App struct {
	*Teardown

	// version is the application version.
	Version string

	// frontendPath is an absolute path to a runtime frontend. When
	// empty will use embedded frontend.
	frontendPath string

	// livePath is the absolute path to the live DDT.
	livePath string

	// componentsPath is an absolute path to a directory containing (transpiled and bundled) assets of a component library
	componentsPath string

	LiveConfigDB config.DB

	Broker *bus.Broker

	Watcher *notify.Watcher

	Sources *Sources

	Components *Components

	Frontend *frontend.Frontend
}

func (app *App) Open() error {
	log.Print("Preparing live source...")

	b, err := bus.NewBroker()
	if err != nil {
		return err
	}
	app.Broker = b
	app.Teardown.AddFunc(b.Close)

	w, err := notify.NewWatcher(app.livePath)
	if err != nil {
		return err
	}
	app.Watcher = w
	app.Teardown.AddFunc(w.Close)

	ok, f, err := config.FindFile(app.livePath)
	if err != nil {
		return err
	}
	if ok {
		cdb, err := config.NewFileDB(
			f,
			norm.NFC.String(filepath.Base(app.livePath)),
		)
		if err != nil {
			return err
		}
		app.LiveConfigDB = cdb
		app.Teardown.AddFunc(app.LiveConfigDB.Close)

		done := app.Watcher.SubscribeFunc("fs.changed", app.LiveConfigDB.Refresh)
		app.Teardown.AddChan(done)
	} else {
		app.LiveConfigDB = config.NewStaticDB(
			norm.NFC.String(filepath.Base(app.livePath)),
		)
	}

	if app.frontendPath == "" {
		app.Frontend = frontend.NewFrontendFromEmbedded(app.livePath)
	} else {
		frontend, err := frontend.NewFrontendFromPath(app.frontendPath, app.livePath)
		if err != nil {
			return err
		}
		app.Frontend = frontend
	}

	ss, err := NewSources(app.LiveConfigDB)
	if err != nil {
		return err
	}
	app.Sources = ss
	app.Teardown.AddFunc(ss.Close)

	live, err := app.Sources.Add("live", app.livePath)
	if err != nil {
		return err
	}

	// We'd like to pull all messages from the source into our main
	// broker.
	done := live.ReverseConnectBroker(app.Broker)
	app.Teardown.AddChan(done)

	log.Print("Successfully prepared live source")

	if app.LiveConfigDB.IsAcceptedSource("live") {
		if err := app.Sources.SelectPrimary("live"); err != nil {
			return err
		}
	}

	if !app.HasMultiVersionsSupport() {
		log.Print("No multi versions support enabled")
		// "live" may not be whitelisted, but without multi versions
		// support that configuration is not honored.
		return app.Sources.SelectPrimary("live")
	}

	// Pre-register the sources with their name, so the sources struct
	// represents a complete (however not entirely initialized list).
	versions, err := live.Versions()
	if err != nil {
		return err
	}

	// We don't want to setup all possible versions.
	versions = versions.Filter(func(v *vcs.Version) bool {
		return app.LiveConfigDB.IsAcceptedSource(v.Name)
	})

	// The DDT might be stored inside a subpath, we assume
	// that it is stored in the same subpath in all versions.
	subpath, _ := filepath.Rel(live.Repo.Path, app.livePath)
	if subpath != "." {
		log.Printf("Detected DDT in subpath: %s", subpath)
	}

	// All versions are lazily completed, by cloning from the live
	// repository. The teardown must happen on the source requesting
	// the clone, to ensure it doesn't block.
	err = versions.ForEach(func(v *vcs.Version) error {
		source, err := app.Sources.AddLazy(v.Name, func(s *Source) (string, *git.Repository, error) {
			return s.CloneFrom(live.Repo.Path, v.Ref.Name())
		})

		// We'd like to pull all messages from the source into our main
		// broker.
		done := source.ReverseConnectBroker(app.Broker)
		app.Teardown.AddChan(done)

		live.Broker.SubscribeFunc("repo.changed", func() error {
			if source.IsComplete() {
				return source.UpdateFromUpstream()
			}
			return nil
		})

		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (app *App) HasMultiVersionsSupport() bool {
	ok, s, _ := app.Sources.Get("live")
	if !ok {
		return false
	}
	return s.HasMultiVersionsSupport()
}

func (app *App) OpenVersions(ctx context.Context) error {
	go app.Sources.ForEach(func(s *Source) error {
		select {
		case <-ctx.Done():
			return nil // Stop
		default:
			// Continue normally
		}
		if s.IsComplete() != true {
			return s.Complete()
		}
		return nil
	})
	return nil
}

func (app *App) OpenComponents(ctx context.Context) error {
	cmps, err := NewComponents(app.componentsPath)
	if err != nil {
		return err
	}

	cmps.Detect()
	app.Components = cmps
	return err
}

func (app *App) Close() error {
	return app.Teardown.Close()
}
