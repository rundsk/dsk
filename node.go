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
	"mime"
	"os"
	"path/filepath"
	"regexp"
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
	Files   []FileInfo
}

// Meta data as specified in a node configuration file.
type NodeMeta struct {
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
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
	n.Files, err = n.filesForNode()

	if n.URL != "/" {
		n.Title = n.titleForUrl(n.URL)
	}
	//log.Printf(n.Files)
	return err
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
func (n Node) Asset(name string) (bytes.Buffer, string, error) {
	var b bytes.Buffer

	c, err := ioutil.ReadFile(filepath.Join(n.path, name))
	if err != nil {
		return b, "", err
	}

	typ := mime.TypeByExtension(filepath.Ext(name))

	b.Write(c)
	return b, typ, nil
}

type FileInfo struct {
	Name  string
	Size  int64
	Mode  os.FileMode
	IsDir bool
	Path  string
	Type  string
}

// Checks node's directory for files
func (n Node) filesForNode() ([]FileInfo, error) {
	files, err := ioutil.ReadDir(n.path)

	if err != nil {
		return nil, err
	}

	filteredFiles := []FileInfo{}

	for _, entry := range files {
		var name string
		name = entry.Name()
		path := filepath.Join(n.path, name)
		filetype := filepath.Ext(name)

		if err != nil {
			return nil, err
		}

		f := FileInfo{
			Name:  entry.Name(),
			Size:  entry.Size(),
			Mode:  entry.Mode(),
			IsDir: entry.IsDir(),
			Path:  path,
			Type:  filetype,
		}

		if f.IsDir != true &&
			f.Name != "readme.md" &&
			f.Name != "api.md" &&
			f.Name != ".DS_Store" &&
			f.Type != ".json" {
			filteredFiles = append(filteredFiles, f)
		}
	}
	return filteredFiles, nil
}

// Returns an alphabetically sorted list of keywords.
func (n Node) Keywords() []string {
	keywords := n.Meta.Keywords

	sort.Strings(keywords)
	return keywords
}

// Returns the full description of the node. Provided for symmetry of
// the node API. There should be no reason to access .Meta directly
// anymore.
func (n Node) Description() string {
	return n.Meta.Description
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
	return n.titleForUrl(url)
}

func (n Node) titleForUrl(url string) string {
	title := filepath.Base(url)
	re := regexp.MustCompile(`^\d+[_,-]{1}(.*)`)
	s := re.FindStringSubmatch(title)

	if len(s) > 0 {
		title = s[1]
	}

	return title
}
