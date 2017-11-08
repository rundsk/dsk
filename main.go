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

	"github.com/gamegos/jsend"
)

var (
	Version string
	sigc    chan os.Signal
	// Absolute path to design definitions root direcrtory.
	root string
	// Check for variant suffix i.e. :foo or :foo%20bar.
	demoRouteRegex = regexp.MustCompile(`^(.+):(.+)$`)
)

func main() {
	log.Printf("starting DSK Version %s", Version)

	if len(os.Args) > 2 {
		log.Fatalf("too many arguments given, expecting exactly 0 or 1")
	}

	if len(os.Args) == 2 {
		root = os.Args[1]
	} else {
		root, _ = os.Getwd()
	}
	root, err := filepath.Abs(root)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using %s as root directory", root)

	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("caught %v signal, bye!", sig)
			// implement cleanup when necessary
			os.Exit(1)
		}
	}()

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	flag.Parse()

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("listening on %s", addr)

	log.Printf("DSK started, please go to: http://%s", addr)
	log.Print("hit STRG+C to quit")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/assets/", assetsHandler)
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/tree/", nodeHandler)
	http.HandleFunc("/embed/", embedHandler)

	http.ListenAndServe(addr, nil)
}

// The root page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("index.html")

	tVars := struct {
		ProjectName string
	}{
		ProjectName: filepath.Base(root),
	}
	html, err := Asset("data/views/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = t.Parse(string(html[:]))
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, tVars); err != nil {
		log.Fatal(err)
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
	path := r.URL.Path[len("/api/"):]

	// Security: Although path is not yet used for file access, we check it, to prevent
	// programmer mistakenly opening a security hole when the code section below is expanded
	// and the path then used.
	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wr := jsend.Wrap(w)

	if path == "tree" {
		nodeList, _ := NewNodeListFromPath(root)
		data := struct {
			NodeList []*Node `json:"nodeList"`
		}{
			NodeList: nodeList,
		}
		wr.
			Data(data).
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
	path := filepath.Join(root, r.URL.Path[len("/tree/"):])

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := NewNodeFromPath(path, root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t := template.New("node.html")

	tVars := struct {
		N *Node
	}{
		N: n,
	}
	html, err := Asset("data/views/node.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = t.Parse(string(html[:]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, tVars); err != nil {
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
	path := filepath.Join(root, r.URL.Path[len("/embed/"):])

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = strings.TrimSuffix(path, ".css")

	n, err := NewNodeFromPath(path, root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	path := filepath.Join(root, r.URL.Path[len("/embed/"):])

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = strings.TrimSuffix(path, ".js")

	n, err := NewNodeFromPath(path, root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	path := filepath.Join(root, r.URL.Path[len("/embed/"):])

	if err := checkSafePath(path, root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var propSet PropSet
	var n *Node
	var err error

	if m := demoRouteRegex.FindStringSubmatch(path); m != nil {
		n, err = NewNodeFromPath(m[1], root)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		propSet, _ = n.Demo(m[2]) // Is auto-unescaped.
	} else {
		n, err = NewNodeFromPath(path, root)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	t := template.New("stage.html")

	mPropSet, _ := json.Marshal(propSet)

	tVars := struct {
		N       *Node
		PropSet template.JS // marshalled PropSet
	}{
		N:       n,
		PropSet: template.JS(mPropSet),
	}
	html, err := Asset("data/views/stage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = t.Parse(string(html[:]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, tVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
