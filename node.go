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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/russross/blackfriday"
)

// Node represents a directory inside the design definitions tree.
type Node struct {
	path     string
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Children []*Node  `json:"children"`
	Meta     NodeMeta `json:"meta"`
	// Ghosted nodes are nodes that have incomplete information, for
	// these nodes not all methods are guaranteed to succeed.
	IsGhost bool `json:"isGhost"`
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

// Constructs a new node using its path in the filesystem. Returns a
// node instance even if errors happened. In which case the node will
// be flagged as "ghost" node.
//
// The URL of each node should end with a trailing slash as to allow
// contained assets to references it as if it was a directory.
func NewNodeFromPath(path string, root string) (*Node, error) {
	var url string
	if path == root {
		url = "/"
	} else {
		url = strings.TrimPrefix(path, root+"/") + "/"
	}

	n := &Node{
		path: path,
		URL:  url,
		// Initialize, so JSON marshalling turns this into `[]` instead of
		// `null` for easier iteration.
		Children: []*Node{},
		Title:    filepath.Base(path),
		IsGhost:  true,
	}
	return n, n.Sync()
}

// One way sync: update node meta data from file system.
func (n *Node) Sync() error {
	meta, err := n.parseNodeConfig()
	if err != nil {
		return err
	}
	n.Meta = meta
	n.IsGhost = false
	return nil
}

// Reads node configuration file when present and returns values. When file
// is not present will simply return an empty Meta.
func (n *Node) parseNodeConfig() (NodeMeta, error) {
	var meta NodeMeta
	f := filepath.Join(n.path, ConfigBasename)

	if _, err := os.Stat(f); os.IsNotExist(err) {
		return meta, nil
	}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(content, &meta); err != nil {
		return meta, fmt.Errorf("failed parsing %s: %s", prettyPath(f), err)
	}
	return meta, nil
}

// Checks node's directory for given asset.
func (n Node) Asset(name string) (bytes.Buffer, error) {
	var b bytes.Buffer

	files, err := filepath.Glob(filepath.Join(n.path, name))
	if err != nil {
		return b, err
	}
	if len(files) == 0 {
		return b, fmt.Errorf("no .%s assets in path %s", n.path, name)
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

// Returns a list of crumb URLs (relative to root). The last element is
// the current. active one. Does not include a root element.
func (n Node) CrumbURLs() []string {
	var urls []string

	parts := strings.Split(strings.TrimSuffix(n.URL, "/"), "/")
	for index, _ := range parts {
		urls = append(urls, strings.Join(parts[:index+1], "/")+"/")
	}
	return urls
}

// Returns the name for a given crumb URL.
func (n Node) CrumbName(url string) string {
	return filepath.Base(url)
}

func (n Node) HasDemos() bool {
	return len(n.Meta.Demos) > 0
}

// Returns the names of all available demos in order. The prop set of
// each demo can be retrieved via Demo(). This approach is needed as
// Go's maps are not guaranteed to keep order.
func (n Node) DemoNames() []string {
	var names []string

	for name := range n.Meta.Demos {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Access a node's demo by its name.
func (n Node) Demo(name string) (PropSet, error) {
	if val, ok := n.Meta.Demos[name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("no demo with name: %s", name)
}
