// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

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

	// Global instance of the search index.
	search *Search
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
				tree.Close()
			}
			if watcher != nil {
				watcher.Close()
			}
			if search != nil {
				search.Close()
			}
			if broker != nil {
				broker.Close()
			}
			os.Exit(1)
		}
	}()

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	version := flag.Bool("version", false, "print DSK version")
	noColor := flag.Bool("no-color", false, "disables color output")
	flang := flag.String("lang", "en", "language; separate multiple languages by commas")
	flag.Parse()

	// Used for configuring search.
	langs := strings.Split(*flang, ",")

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

	log.Print("Starting message broker...")
	broker = NewMessageBroker() // assign to global
	broker.Start()

	log.Printf("Detecting tree root...")
	here, err := detectTreeRoot(os.Args[0], flag.Arg(0))
	if err != nil {
		log.Fatal(red.Sprintf("Failed to detect root of design definitions tree: %s", err))
	}

	log.Printf("Tree root found: %s", here)
	PrettyPathRoot = here

	log.Print("Begin watching tree for changes...")
	w := NewWatcher(here)
	if err := w.Open(IgnoreNodesRegexp); err != nil {
		log.Fatal(red.Sprintf("Failed to install watcher: %s", err))
	}
	watcher = w // assign to global

	log.Print("Opening tree...")
	tree = NewNodeTree(here, watcher, broker) // assign to global

	if err := tree.Open(); err != nil {
		log.Fatal(red.Sprintf("Failed to open tree: %s", err))
	}
	if err := tree.Sync(); err != nil {
		log.Fatal(red.Sprintf("Failed to perform initial tree sync: %s", err))
	}

	log.Print("Opening search index...")
	search = NewSearch(tree, broker, langs) // assign to global
	if err := search.Open(); err != nil {
		log.Fatal(red.Sprintf("Failed to open search index: %s", err))
	}

	if err := search.IndexTree(); err != nil {
		log.Fatal(red.Sprintf("Failed to perform initial tree indexing: %s", err))
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
	log.Print("Mounting frontend...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			assetHandler(w, r)
		} else {
			rootHandler(w, r)
		}
	})

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Starting web interface on %s....", addr)

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
func rootHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, ""}
	path := "index.html"

	// Does not check on path, as we only ever serve a single
	// file from here, and that path is hard-coded.

	buf, err := Asset(path)
	if err != nil {
		wr.Error(HTTPErrNoSuchAsset, err)
		return
	}
	info, err := AssetInfo(path)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(buf))
}

// Serves the frontend's assets.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
func assetHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, ""}
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, tree.path); err != nil {
		wr.Error(HTTPErrUnsafePath, err)
		return
	}

	buf, err := Asset(path)
	if err != nil {
		wr.Error(HTTPErrNoSuchAsset, err)
		return
	}
	info, err := AssetInfo(path)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(buf))
}
