// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
)

var (
	// Version string, compiled in.
	Version string

	// OS Signal channel.
	sigc chan os.Signal

	// Instance of the design defintions tree.
	tree *NodeTree

	// Watcher instance overseeing the tree for changes.
	watcher *Watcher

	// Global instance of a message broker.
	broker *MessageBroker

	// Global instance of the repository.
	repository *Repository

	// Global instance of the search index.
	search *Search

	// Global instance of a http.Filesystem to access frontend assets
	frontend http.FileSystem
)

func main() {
	// Disable prefix, we are invoked directly.
	log.SetFlags(0)
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	// Listen for interrupt and allow to cancel program early.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v signal, bye!", sig)
			log.Print("Cleaning up...")

			// Close services in reverse order of starting them. They might
			// not yet have been started, if we've been invoked early.
			if tree != nil {
				tree.StopSyncer()
				// tree doesn't need to be closed.
			}
			if watcher != nil {
				watcher.Stop()
				watcher.Close()
			}
			if repository != nil {
				repository.StopLookupBuilder()
				repository.Close()
			}
			if search != nil {
				search.StopIndexer()
				search.Close()
			}
			if broker != nil {
				broker.Stop()
				broker.Close()
			}
			os.Exit(1)
		}
	}()

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	version := flag.Bool("version", false, "print DSK version")
	noColor := flag.Bool("no-color", false, "disables color output")
	flang := flag.String("lang", "en", "language the documents are authored in")
	ffrontend := flag.String("frontend", "", "path to a frontend, to use instead of the built-in")
	flag.Parse()

	if len(flag.Args()) > 1 {
		log.Fatalf("Too many arguments given, expecting exactly 0 or 1")
	}

	if *version {
		fmt.Println(Version)
		os.Exit(1)
	}

	// Color package automatically disables colors when not a TTY. We
	// don't need to check for an interactive terminal here again.
	if *noColor {
		color.NoColor = true
	}
	whiteOnBlue := color.New(color.FgWhite, color.BgBlue)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	if isTerminal {
		log.Print(whiteOnBlue.Sprint(" DSK "))
		log.Printf("Version %s", Version)
		log.Print()
	}

	here, err := detectTreeRoot(os.Args[0], flag.Arg(0))
	if err != nil {
		log.Fatal(red.Sprintf("Failed to detect root of design definitions tree: %s", err))
	}

	log.Printf("Tree root found: %s", here)
	PrettyPathRoot = here

	authors := NewAuthors(here)
	broker = NewMessageBroker() // assign to global
	watcher = NewWatcher(here)  // assign to global

	rroot, err := detectRepository(here, false)
	if err != nil && err != ErrRepositoryNotFound {
		log.Fatal(red.Sprintf("Failed to detect repository: %s", err))
	}
	if err != ErrRepositoryNotFound {
		log.Printf("Detected VCS support in: %s", rroot)

		rsub, err := detectRepository(here, true)
		if err == nil {
			log.Printf("Using submodule in: %s", rsub)
		}
		if err != nil && err != ErrRepositoryNotFound {
			log.Fatal(red.Sprintf("Failed to detect repository: %s", err))
		}

		repository, err = NewRepository(rroot, rsub) // assign to global
		if err != nil {
			log.Fatal(red.Sprintf("Failed to enable VCS support: %s", err))
		}
	}
	tree = NewNodeTree(here, authors, repository, watcher, broker) // assign to global
	search, err = NewSearch(tree, broker, *flang)                  // assign to global
	if err != nil {
		log.Fatal(red.Sprintf("Failed to open search index: %s", err))
	}

	broker.Start()

	if err := watcher.Start(); err != nil {
		log.Fatal(red.Sprintf("Failed to start watcher: %s", err))
	}
	tree.StartSyncer()
	if repository != nil {
		repository.StartLookupBuilder()
	}
	search.StartIndexer()

	if err := authors.Sync(); err != nil {
		log.Fatal(red.Sprintf("Failed to perform initial authors sync: %s", err))
	}
	if err := tree.Sync(); err != nil {
		log.Fatal(red.Sprintf("Failed to perform initial tree sync: %s", err))
	}
	// tree.Sync() will indirectly through messaging trigger initial indexing.
	if repository != nil {
		go func() {
			if err := repository.BuildLookup(); err != nil {
				log.Fatal(red.Sprintf("Failed to build initial repository cache: %s", err))
			}
		}()
	}

	if *ffrontend != "" {
		*ffrontend, err = filepath.Abs(*ffrontend)
		if err != nil {
			log.Fatal(red.Sprintf("Failed using frontend provided by path: %s", err))
		}
		frontend = http.Dir(*ffrontend) // assign to global
		log.Printf("Using runtime frontend from: %s", prettyPath(*ffrontend))
	} else {
		frontend = assets // assign to global, using built-in
		log.Print("Using built-in frontend")
	}

	apis := map[int]API{
		1: NewAPIv1(tree, broker, search),
		2: NewAPIv2(tree, broker, search),
	}
	for v, api := range apis {
		log.Printf("Mounting APIv%d...", v)
		api.MountHTTPHandlers()
	}

	// Handles frontend root document delivery and frontend assets.
	// The frontend is allowed to use any path except /api. We route
	// everything else into the front controller (index.html).
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			frontendAssetHandler(w, r)
		} else {
			frontendRootHandler(w, r)
		}
	})

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Starting web interface on %s...", addr)

	if isTerminal {
		log.Print()
		log.Printf("Please visit: %s", green.Sprint("http://"+addr))
		log.Print("Hit Ctrl+C to quit")
		log.Print()
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(red.Sprintf("Failed to start web interface: %s", err))
	}
}

// Serves the frontend's index.html.
//
// Handles these kinds of URLs:
//   /
//   /index.html
//   /* <catch all>
func frontendRootHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, ""}
	path := "index.html"

	// Does not check on path, as we only ever serve a single
	// file from here, and that path is hard-coded.

	asset, err := frontend.Open(path)
	if err != nil {
		wr.Error(HTTPErrNoSuchAsset, err)
		return
	}
	defer asset.Close()

	info, err := asset.Stat()
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), asset)
}

// Serves the frontend's assets.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
func frontendAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, ""}
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, tree.path); err != nil {
		wr.Error(HTTPErrUnsafePath, err)
		return
	}

	asset, err := frontend.Open(path)
	if err != nil {
		wr.Error(HTTPErrNoSuchAsset, err)
		return
	}
	defer asset.Close()

	info, err := asset.Stat()
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), asset)
}
