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
)

var (
	Version string

	sigc chan os.Signal

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

	here, err := detectRoot(os.Args[0], flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to detect root of design definitions tree: %s", red(err))
	}
	PrettyPathRoot = here
	log.Printf("Reading design definitions tree from %s...", prettyPath(here))

	tree = NewNodeTree(here) // assign to global
	if err := tree.Sync(); err != nil {
		log.Fatalf("Failed to do initial tree sync: %s", red(err))
	}
	if err := tree.StartAutoSync(); err != nil {
		log.Fatalf("Failed to start tree auto-sync: %s", red(err))
	}

	log.Print("Mounting APIv1...")
	apiv1 := &APIv1{tree}
	apiv1.MountHTTPHandlers(withNoop)

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
	log.Printf("Will listen on %s", addr)

	log.Printf("Please visit: %s", green("http://"+addr))
	log.Print("Hit Ctrl+C to quit")

	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v signal, bye!", sig)
			log.Print("Cleaning up...")
			tree.StopAutoSync()
			os.Exit(1)
		}
	}()

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

	data, err := Asset("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(data[:])
}

// Serves the frontend's assets.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
func assetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, tree.path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	buf, err := Asset(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	typ := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", typ)
	w.Write(buf[:])
	return
}
