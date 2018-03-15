// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ConfigBasename = "index.json"
)

var (
	// Basenames matching the pattern will be ignored when searching
	// for downloadable files in the node's directory.
	IgnoreDownloadsRegexp = regexp.MustCompile(`(?i)^(.*\.(js|css|md|markdown|json)|\.DS_Store|\.git.*|dsk)$`)

	// Basenames matching this pattern are considered documents.
	NodeDocsRegexp = regexp.MustCompile(`(?i)^.*\.(md|markdown)$`)

	// Patterns for extracting order number and title from a node's
	// path/URL segment in the form of 06_Foo. As well as for
	// "slugging" the URL/path segment.
	NodePathTitleRegexp        = regexp.MustCompile(`^0?(\d+)[_,-]+(.*)$`)
	NodePathInvalidCharsRegexp = regexp.MustCompile("[^A-Za-z0-9-_]")
	NodePathMultipleDashRegexp = regexp.MustCompile("-+")
)

// Constructs a new synced node using its path in the filesystem.
// Returns a node instance even if uncritical errors happened. In that
// case the node will be flagged as a "ghost" node.
func NewNodeFromPath(path string, root string) (*Node, error) {
	n := &Node{root: root, path: path, children: []*Node{}}

	m, err := NewNodeMetaFromPath(n.path)
	n.IsGhost = err != nil
	n.meta = m

	return n, nil
}

// Node represents a directory inside the design definitions tree.
type Node struct {
	// Absolute path to the design defintions tree root.
	root string
	// Absolute path to the node's directory.
	path string
	// A list of children nodes. When Node is used in a flat list
	// may be left empty.
	children []*Node
	// Meta data as parsed from the node configuration file.
	meta NodeMeta
	// Ghosted nodes are nodes that have incomplete information, for
	// these nodes not all methods are guaranteed to succeed.
	IsGhost bool
}

// One way sync: update node meta data from file system.
func (n *Node) Sync() error {
	m, err := NewNodeMetaFromPath(n.path)
	n.IsGhost = err != nil
	n.meta = m

	return nil
}

// Returns the normalized URL path fragment, that can be used to
// address this node i.e Input/Password.
func (n Node) URL() string {
	if n.root == n.path {
		return ""
	}
	return normalizeNodeURL(strings.TrimPrefix(n.path, n.root+"/"))
}

// An order number, as a hint for outside sorting mechanisms.
func (n Node) Order() uint64 {
	return orderNumber(filepath.Base(n.path))
}

// Returns the list of children nodes. May be left empty when node is
// used in a flat list of results, where children information is not
// needed.
func (n Node) Children() []*Node {
	return n.children
}

func (n *Node) AddChild(cn *Node) {
	n.children = append(n.children, cn)
}

// The node's computed title with any ordering numbers stripped off, usually for display purposes.
func (n Node) Title() string {
	if n.root == n.path {
		return ""
	}
	return removeOrderNumber(filepath.Base(n.path))
}

// Returns the full description of the node.
func (n Node) Description() string {
	return n.meta.Description
}

// Returns a list of related nodes.
func (n Node) Related() []string {
	if n.meta.Related == nil {
		return make([]string, 0)
	}
	return n.meta.Related
}

// Returns an alphabetically sorted list of tags.
func (n Node) Tags() []string {
	if n.meta.Tags == nil {
		return make([]string, 0)
	}
	tags := n.meta.Tags

	sort.Strings(tags)
	return tags
}

// Returns a list of keywords terms.
func (n Node) Keywords() []string {
	if n.meta.Keywords == nil {
		return make([]string, 0)
	}
	return n.meta.Keywords
}

// Returns a list of node authors; wil use the given authors
// database to lookup information.
func (n Node) Authors(as *Authors) []*Author {
	r := make([]*Author, 0)

	if n.meta.Authors == nil {
		return r
	}
	for _, email := range n.meta.Authors {
		author := as.Get(email)
		if author == nil {
			author = &Author{email, ""}
		}
		r = append(r, author)
	}
	return r
}

// Finds the most recently edited file in the node directory and
// returns its modified timestamp.
func (n Node) Modified() time.Time {
	var modified time.Time

	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return modified
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if f.ModTime().After(modified) {
			modified = f.ModTime()
		}
	}
	return modified
}

func (n Node) Version() string {
	return n.meta.Version
}

// Returns a node asset, given its basename.
func (n Node) Asset(name string) (*NodeAsset, error) {
	path := filepath.Join(n.path, name)

	f, err := os.Stat(path)
	if os.IsNotExist(err) || err != nil {
		return nil, err
	}
	if f.IsDir() {
		return nil, fmt.Errorf("Accessing directory as asset: %s", path)
	}
	return &NodeAsset{
		path: filepath.Join(n.path, f.Name()),
		Name: f.Name(),
		URL:  filepath.Join(n.URL(), f.Name()),
	}, nil
}

// Returns a list of downloadable files, this may include Sketch files
// or other binary assets. JavaScript and Stylesheets and DSK control
// files are excluded.
func (n Node) Downloads() ([]*NodeAsset, error) {
	downloads := make([]*NodeAsset, 0)

	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return downloads, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if IgnoreDownloadsRegexp.MatchString(f.Name()) {
			continue
		}
		downloads = append(downloads, &NodeAsset{
			path: filepath.Join(n.path, f.Name()),
			Name: f.Name(),
			URL:  filepath.Join(n.URL(), f.Name()),
		})
	}
	return downloads, nil
}

// Returns a slice of documents for this node.
//
// The provided prefix will be used to make relative links inside the
// document absolute.
func (n Node) Docs(prefix string) ([]*NodeDoc, error) {
	docs := make([]*NodeDoc, 0)

	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return docs, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !NodeDocsRegexp.MatchString(f.Name()) {
			continue
		}
		docs = append(docs, &NodeDoc{
			path:      filepath.Join(n.path, f.Name()),
			URLPrefix: prefix,
		})
	}
	return docs, nil
}

// Returns a list of crumbs. The last element is the current active
// one. Does not include a root element.
func (n Node) Crumbs() []*NodeCrumb {
	crumbs := make([]*NodeCrumb, 0)

	parts := strings.Split(strings.TrimSuffix(n.URL(), "/"), "/")
	for index, part := range parts {
		crumbs = append(crumbs, &NodeCrumb{
			Title: removeOrderNumber(part),
			URL:   strings.Join(parts[:index+1], "/"),
		})
	}
	return crumbs
}

// "Prettifies" (and normalizes) given relative node URL. Idempotent
// function. Removes any order numbers, as well as leading and
// trailing slashes.
//
//   /foo/bar/  -> foo/bar
//   foo/02_bar -> foo/bar
func normalizeNodeURL(url string) string {
	var normalized []string

	for _, p := range strings.Split(url, "/") {
		if p == "/" {
			continue
		}
		p = removeOrderNumber(p)
		p = NodePathInvalidCharsRegexp.ReplaceAllString(p, "-")
		p = NodePathMultipleDashRegexp.ReplaceAllString(p, "-")
		p = strings.Trim(p, "-")

		normalized = append(normalized, p)
	}
	return strings.Join(normalized, "/")
}

// Finds an order number embedded into given path/URL segment and
// returns it. If none is found, returns 0.
func orderNumber(segment string) uint64 {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) > 2 {
		parsed, _ := strconv.ParseUint(s[0], 10, 64)
		return parsed
	}
	return 0
}

// Removes order numbers from path/URL segment, if present.
func removeOrderNumber(segment string) string {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) == 0 {
		return segment
	}
	if len(s) > 2 {
		return s[2]
	}
	return s[1]
}
