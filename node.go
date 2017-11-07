// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

// Node represents a directory inside the design definitions tree.
type Node struct {
	path     string
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Parent   *Node    `json:"parent"`
	Children []*Node  `json:"children"`
	Meta     NodeMeta `json:"meta"`
}

// Meta data as specified in a node configuration file.
type NodeMeta struct {
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	// Optional, if missing will use the URL.
	Import string
	// Optionally defines a list of property sets, keyed by their names.
	Demos map[string]PropSet `json:"demos"`
}

// A set of component properties, usually parsed from JSON.
type PropSet interface{}

const (
	ConfigBasename     = "index.json"
	GeneralDocBasename = "readme.md"
	APIDocBasename     = "api.md"
)

// Constructs a new node using its path in the filesystem.
func NewNodeFromPath(path string, root string) (*Node, error) {
	var url string

	if path == root {
		url = "/"
	} else {
		url = strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), "/")
	}

	meta, err := parseNodeConfig(path)
	if err != nil {
		return nil, err
	}

	return &Node{
		path:  path,
		URL:   url,
		Title: filepath.Base(path),
		Meta:  meta,
	}, nil
}

// Recursively crawls the given root directory, constructing a flat list
// of nodes.
func NewNodeListFromPath(root string) ([]*Node, error) {
	var nodes []*Node

	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			n, nErr := NewNodeFromPath(path, root)
			if nErr != nil {
				log.Printf("skipping node: %s", nErr)
			} else {
				nodes = append(nodes, n)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build node tree in %s: %s", root, err)
	}

	for _, n := range nodes {
		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
			// TODO: When adding parent, the conversion to JSON will create
			// recursions.
			// if filepath.Dir(n.path) == sn.path {
			// 	n.Parent = sn
			// }
		}
	}

	return nodes, nil
}

// Reads node configuration file when present and returns values. When file
// is not present will simply return an empty Meta.
func parseNodeConfig(path string) (NodeMeta, error) {
	var meta NodeMeta
	f := filepath.Join(path, ConfigBasename)

	if _, err := os.Stat(f); os.IsNotExist(err) {
		return meta, nil
	}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(content, &meta); err != nil {
		return meta, fmt.Errorf("failed parsing %s: %s", f, err)
	}
	return meta, nil
}

// Result is passed as component import name to renderComponent()
// JavaScript glue function.
func (n Node) Import() (string, error) {
	if n.Meta.Import != "" {
		return n.Meta.Import, nil
	}
	return n.URL, nil
}

func (n Node) HasComponent() bool {
	return n.HasJS()
}

func (n Node) HasJS() bool {
	files, _ := filepath.Glob(filepath.Join(n.path, "*.js"))
	return len(files) > 0
}

func (n Node) JS() (bytes.Buffer, error) {
	return n.bundledAssets("js")
}

func (n Node) HasCSS() bool {
	files, _ := filepath.Glob(filepath.Join(n.path, "*.css"))
	return len(files) > 0
}

func (n Node) CSS() (bytes.Buffer, error) {
	return n.bundledAssets("css")
}

// Looks for i.e. CSS files in node directory and concatenates them.
// This way we don't need a naming convention for these assets.
func (n Node) bundledAssets(suffix string) (bytes.Buffer, error) {
	var b bytes.Buffer

	files, err := filepath.Glob(filepath.Join(n.path, "*."+suffix))
	if err != nil {
		return b, err
	}
	if len(files) == 0 {
		return b, fmt.Errorf("no .%s assets in path %s", suffix, n.path)
	}

	for _, f := range files {
		c, err := ioutil.ReadFile(f)
		if err != nil {
			return b, err
		}
		b.Write(c)
	}
	return b, nil
}

// Checks whether general documentation is available.
func (n Node) HasGeneralDoc() bool {
	_, err := os.Stat(filepath.Join(n.path, GeneralDocBasename))
	return !os.IsNotExist(err)
}

// Returns general documentation parsed from markdown into HTML format.
func (n Node) GeneralDoc() (template.HTML, error) {
	contents, err := ioutil.ReadFile(filepath.Join(n.path, GeneralDocBasename))
	if err != nil {
		return template.HTML(""), err
	}
	return template.HTML(blackfriday.Run(contents)), nil
}

// Checks whether API documentation is available.
func (n Node) HasAPIDoc() bool {
	_, err := os.Stat(filepath.Join(n.path, APIDocBasename))
	return !os.IsNotExist(err)
}

// Returns API documentation parsed from markdown into HTML format.
func (n Node) APIDoc() (template.HTML, error) {
	contents, err := ioutil.ReadFile(filepath.Join(n.path, APIDocBasename))
	if err != nil {
		return template.HTML(""), err
	}
	return template.HTML(blackfriday.Run(contents)), nil
}

// Returns a mapping of URLs to title strings for easily creating a
// breadcrumb navigation. The last element is the current active one.
// Does not include the very root element.
func (n Node) Crumbs() map[string]string {
	var crumbs map[string]string // maps url to title
	// TODO
	//	parts := strings.Split(n.url, "/")
	//
	//	for _, p := range parts {
	//		title := p
	//		url :=
	//	}
	return crumbs
}

// Access a node's demo by its name.
func (n Node) Demo(name string) (PropSet, error) {
	if val, ok := n.Meta.Demos[name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("no demo with name: %s", name)
}
