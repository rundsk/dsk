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
	"github.com/gamegos/jsend"
)

var (
	Version string

	sigc chan os.Signal

	// Absolute path to design definitions root directory.
	// TODO: rename to treeRoot
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
	log.Printf("Using %s as root directory", root)

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
			return assetHandler(w, r)
		}
		return rootHandler(w, r)
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start web interface: %s", red(err))
	}
}

// Serves the frontend and a node's assets. Will first
// look into the frontend's path then into the design
// defintions tree path.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
//   /DataEntry/Components/Button/test.png
//   /Button/foo.mp4
//
func assetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO:
	//  - first try frontend assets
	//  - use http.ServeFile(w, r, assetPath), to fix video serving
	//  - consider not retrieving node assets via tree, but directly via FS

	// typ := mime.TypeByExtension(filepath.Ext(path))
	// w.Header().Set("Content-Type", typ)

	// data, err := Asset(path)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	return
	// }
	// w.Write(data[:])

	n, err := tree.Get(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	buf, typ, err := n.Asset(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", typ)
	w.Write(buf.Bytes())
	return
}

// Returns the entire tree. Will not get or check path, only tree
// requests are routed here.
//
// Handles this URL:
//   /api/v1/tree
func treeHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)

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

// Renders a given HTML fragment for given node.
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
