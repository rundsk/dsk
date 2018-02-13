// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/gamegos/jsend"
)

var (
	Version string

	sigc chan os.Signal

	// Absolute path to design definitions root directory.
	root string

	// Instance of the design defintions tree.
	tree *NodeTree
)

func main() {
	log.SetFlags(0) // disable prefix, we are invoked directly.

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	noColor := flag.Bool("no-color", false, "disables color output")
	flag.Parse()

	if len(flag.Args()) > 1 {
		log.Fatalf("Too many arguments given, expecting exactly 0 or 1")
	}

	// Color package automatically disables colors when not a TTY. We
	// don't need to check for an interactive terminal here again.
	if *noColor {
		color.NoColor = true
	}
	whiteOnBlue := color.New(color.FgWhite, color.BgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	log.Printf("Starting %s Version %s", whiteOnBlue(" DSK "), Version)

	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v signal, bye!", sig)
			os.Exit(1)
		}
	}()

	here, err := detectRoot(os.Args[0], flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to detect root of design definitions tree: %s", red(err))
	}
	root = here // assign to global
	log.Printf("Using design definitions tree in %s...", prettyPath(root))

	tree = NewNodeTreeFromPath(here) // assign to global
	if err := tree.Sync(); err != nil {
		log.Fatalf("Failed to do initial tree sync: %s", red(err))
	}
	log.Printf("Synced tree with %d total nodes", tree.TotalNodes())

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Will listen on %s", addr)

	log.Printf("Please visit: %s", green("http://"+addr))
	log.Print("Hit Ctrl+C to quit")

	http.HandleFunc("/api/v1/tree", treeHandler)
	http.HandleFunc("/api/v1/tree/", nodeHandler)
	http.HandleFunc("/api/v1/search", searchHandler)

	// Anything that doesn't look like a node or frontend asset, will
	// be routed into the root handler.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			assetHandler(w, r)
		} else {
			rootHandler(w, r)
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start web interface: %s", red(err))
	}
}

// Serves the frontend's index.html.
//
// Handles these kinds of URLs:
//   /
//   /index.html
//   /* <catch all>
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Does not check on path, as we only ever serve a single
	// file from here, and that path is hard-coded.
	w.Header().Set("Content-Type", "text/html")

	data, err := Asset("frontend/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(data[:])
}

// Serves the frontend and a node's assets. Will first look into the
// frontend's path then into the design defintions tree path.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
//   /DataEntry/Components/Button/test.png
//   /Button/foo.mp4
func assetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// First check if this is a frontend asset. There is no way to
	// check if an asset is actually embeded without masking other
	// errors. As these errors are deemed to be seldom enough,
	// we don't care.
	buf, err := Asset(filepath.Join("frontend", path))
	if err == nil {
		typ := mime.TypeByExtension(filepath.Ext(path))
		w.Header().Set("Content-Type", typ)
		w.Write(buf[:])
		return
	}

	n, err := tree.Get(filepath.Dir(path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	assetPath, err := n.Asset(filepath.Base(path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, assetPath)
	return
}

// Returns all nodes in the design defintions tree, as nested nodes.
//
// Handles this URL:
//   /api/v1/tree
func treeHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	// Not getting or checking path here, as only tree requests are routed
	// here.

	if err := tree.Sync(); err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	wr.
		Data(tree).
		Status(http.StatusOK).
		Send()
}

// Returns information about a single node.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table
//   /api/v1/tree/DisplayData/Table/Row
func nodeHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := tree.GetSynced(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	wr.
		Data(n).
		Status(http.StatusOK).
		Send()
}

// Performs a full text search over the design defintions tree and
// returns results.
//
// Handles these kinds of URLs:
//   /api/v1/search?q={query}
//
// TODO: Implement :)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	// path := r.URL.Path[len("/api/v1/search"):]
	wr.
		Status(http.StatusNotImplemented).
		Message("WIP").
		Send()
}
