// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

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

	// Check for variant suffix i.e. :foo or :foo%20bar.
	demoRouteRegex = regexp.MustCompile(`^(.+):(.+)$`)

	// Handful of pre-parsed templates.
	templateIndex = mustPrepareTemplate("index.html")
	templateNode  = mustPrepareTemplate("node.html")
	templateStage = mustPrepareTemplate("stage.html")
)

func main() {
	log.SetFlags(0) // disable prefix, we are invoked directly.

	if len(os.Args) > 2 {
		log.Fatalf("Too many arguments given, expecting exactly 0 or 1")
	}

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	noColor := flag.Bool("no-color", false, "disables color output")
	flag.Parse()

	// Color package automatically disables colors when not a TTY. We
	// don't need to check for an interactive terminal here again.
	if *noColor {
		color.NoColor = true
	}
	whiteOnBlue := color.New(color.FgWhite, color.BgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	log.Printf("Starting %s Version %s", whiteOnBlue(" DSK "), Version)

	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v signal, bye!", sig)
			os.Exit(1)
		}
	}()

	here, err := detectRoot()
	if err != nil {
		log.Fatal(err)
	}
	root = here // assign to global
	log.Printf("Using %s as root directory", root)

	tree = NewNodeTreeFromPath(here) // assign to global
	if err := tree.Sync(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Synced tree with %d total nodes", tree.TotalNodes())

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Listening on %s", addr)

	log.Printf("Please visit: %s", green("http://"+addr))
	log.Print("Hit Ctrl+C to quit")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/assets/", assetsHandler)
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/tree/", nodeHandler)
	http.HandleFunc("/embed/", embedHandler)

	http.ListenAndServe(addr, nil)
}

// The root page.
//
// Handles these kinds of URLs:
//   /DataEntry/Components/Button
//   /DataEntry/Components/Button/
//   /DataEntry/Components/Button/test.png
func indexHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the path contains a file, if yes strip file
	// from path, leaving just the node path.
	var file string
	if filepath.Ext(path) != "" {
		file = filepath.Base(path)
		path = filepath.Dir(path) + "/"
	}

	n, err := tree.Get(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// If the path contained a file, return the file
	if file != "" {
		buf, typ, err := n.Asset(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", typ)
		w.Write(buf.Bytes())
		return
	}

	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, fmt.Sprintf("%s/?%s", r.URL.Path, r.URL.RawQuery), 302)
		return
	}

	tVars := struct {
		ProjectName string
	}{
		ProjectName: filepath.Base(root),
	}
	if err := templateIndex.Execute(w, tVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handles requests for non-node assets.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /assets/js/index.js
func assetsHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("data/assets", r.URL.Path[len("/assets/"):])

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	typ := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", typ)

	data, err := Asset(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(data[:])
}

// Returns JSEND response. Currently only used for returning
// information about the tree for the left hand side navigation:
//
// Handles this URL:
//   /api/tree
func apiHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	path := r.URL.Path[len("/api/"):]

	// Security: Although path is not yet used for file access, we check it, to prevent
	// programmer mistakenly opening a security hole when the code section below is expanded
	// and the path then used.
	if err := checkSafePath(path, root); err != nil {
		wr.
			Status(http.StatusBadRequest).
			Message(err.Error()).
			Send()
		return
	}

	if path == "tree" {
		if err := tree.Sync(); err != nil {
			wr.
				Status(http.StatusInternalServerError).
				Message(err.Error()).
				Send()
			return
		}
		wr.
			Data(tree).
			Status(201).
			Send()
		return
	}
	wr.
		Status(404).
		Send()
}

// Renders a page for given node.
//
// Handles these kinds of URLs:
//   /tree/DisplayData/Table
//   /tree/DisplayData/Table/Row
func nodeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/tree/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", 302)
		return
	}

	n, err := tree.GetSynced(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tVars := struct {
		N *Node
	}{
		N: n,
	}
	if err := templateNode.Execute(w, tVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Renders a HTML page that will be rendered into the components
// stage. It will be embeded using an iframe.
//
// Handles these kinds of URLs:
//   /embed/DisplayData/Table
//   /embed/DisplayData/Table:foo%20bar
//   /embed/DisplayData/Table:bar
//   /embed/DisplayData/Table.js
//   /embed/DisplayData/Table.css
func embedHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") {
		embedHandlerCSS(w, r)
	} else if strings.HasSuffix(r.URL.Path, ".js") {
		embedHandlerJS(w, r)
	} else {
		embedHandlerDemo(w, r)
	}
}

func embedHandlerCSS(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/embed/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = strings.TrimSuffix(path, ".css")

	n, err := tree.GetSynced(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	buf, err := n.CSS()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "text/css")
	w.Write(buf.Bytes())
}

func embedHandlerJS(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/embed/"):]

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = strings.TrimSuffix(path, ".js")

	n, err := tree.GetSynced(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	buf, err := n.JS()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/javascript")
	w.Write(buf.Bytes())
}

func embedHandlerDemo(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/embed/"):]

	var propSet PropSet
	var demo string

	if m := demoRouteRegex.FindStringSubmatch(path); m != nil {
		path = m[1]
		demo = m[2] // Is auto-unescaped.
	}

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := tree.GetSynced(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if demo != "" {
		propSet, _ = n.Demo(demo)
	}
	mPropSet, _ := json.Marshal(propSet)

	tVars := struct {
		N       *Node
		PropSet template.JS // marshalled PropSet
	}{
		N:       n,
		PropSet: template.JS(mPropSet),
	}
	if err := templateStage.Execute(w, tVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
