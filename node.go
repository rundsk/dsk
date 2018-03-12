// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

const (
	ConfigBasename = "index.json"
)

var (
	// Basenames matching the pattern will be ignored when searching
	// for downloadable files in the node's directory.
	IgnoreDownloadsRegexp = regexp.MustCompile(`(?i)^(.*\.(js|css|md|markdown|json)|\.DS_Store|\.git.*|dsk)$`)

	// A pattern for extracting order number and title from a title in the form of 06_Foo.
	NodeTitleRegexp = regexp.MustCompile(`^0?(\d+)[_,-]+(.*)$`)

	// Basenames matching this pattern are considered documents.
	NodeDocsRegexp = regexp.MustCompile(`(?i)^.*\.(md|markdown)$`)
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
	title := filepath.Base(n.path)
	s := NodeTitleRegexp.FindStringSubmatch(title)

	if len(s) > 2 {
		parsed, _ := strconv.ParseUint(s[0], 10, 64)
		return parsed
	}
	return 0
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

// Returns an alphabetically sorted list of tags.
func (n Node) Tags() []string {
	tags := n.meta.Tags

	sort.Strings(tags)
	return tags
}

// Returns a list of keywords terms.
func (n Node) Keywords() []string {
	return n.meta.Keywords
}

// Returns a list of node owners; wil use the given authors
// database to lookup information.
func (n Node) Owners(as *Authors) []*Author {
	var r []*Author

	for _, email := range n.meta.Owners {
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
	results := []*NodeAsset{}

	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return results, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if IgnoreDownloadsRegexp.MatchString(f.Name()) {
			continue
		}
		results = append(results, &NodeAsset{
			path: filepath.Join(n.path, f.Name()),
			Name: f.Name(),
			URL:  filepath.Join(n.URL(), f.Name()),
		})
	}
	return results, nil
}

// Returns a slice of documents for this node.
//
// The provided prefix will be used to make relative links inside the
// document absolute.
func (n Node) Docs(prefix string) ([]*NodeDoc, error) {
	var docs []*NodeDoc

	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return docs, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if NodeDocsRegexp.MatchString(f.Name()) {
			docs = append(docs, &NodeDoc{
				path:      filepath.Join(n.path, f.Name()),
				Name:      f.Name(),
				URLPrefix: prefix,
			})
		}
	}
	return docs, nil
}

// Returns a list of crumbs. The last element is the current active
// one. Does not include a root element.
func (n Node) Crumbs() []*NodeCrumb {
	crumbs := []*NodeCrumb{}

	parts := strings.Split(strings.TrimSuffix(n.URL(), "/"), "/")
	for index, part := range parts {
		crumbs = append(crumbs, &NodeCrumb{
			Title: removeOrderNumber(part),
			URL:   strings.Join(parts[:index+1], "/"),
		})
	}
	return crumbs
}

// Looks for a node configuration file in given directory, parses the
// file and returns a filled NodeMeta struct. If not file is found
// returns an empty NodeMeta.
func NewNodeMetaFromPath(path string) (NodeMeta, error) {
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
		return meta, fmt.Errorf("Failed parsing %s: %s", prettyPath(f), err)
	}
	return meta, nil
}

// Metadata parsed from node configuration.
type NodeMeta struct {
	Description string
	Keywords    []string
	Tags        []string
	Owners      []string // Email addresses of node owners.
	Version     string   // Freeform version string.
}

// A markdown document file.
type NodeDoc struct {
	// Absolute path to the file.
	path string
	// The basename of the file, usually for display purposes.
	Name string
	// The provided prefix will be used to make relative links inside the
	// document absolute.
	URLPrefix string
}

// Raw content of the underlying file.
func (d NodeDoc) Raw() ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// HTML as parsed from the underlying file.
func (d NodeDoc) HTML() ([]byte, error) {
	switch filepath.Ext(d.path) {
	case ".md", ".markdown":
		return d.parseMarkdown()
	}
	return nil, fmt.Errorf("document %s is not in a supported format", d.path)
}

func (d NodeDoc) parseMarkdown() ([]byte, error) {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags:          blackfriday.CommonHTMLFlags &^ blackfriday.UseXHTML,
		AbsolutePrefix: d.URLPrefix,
	})
	extensions := blackfriday.CommonExtensions |
		blackfriday.Strikethrough | blackfriday.NoEmptyLineBeforeBlock&^
		blackfriday.HeadingIDs&^blackfriday.DefinitionLists

	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	return blackfriday.Run(
		contents,
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(extensions),
	), nil
}

// A downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	path string
	// The URL, relative to the design defintion tree root.
	URL string
	// The basename of the file, usually for display purposes.
	Name string
}

type NodeCrumb struct {
	URL   string
	Title string
}

// Normalizes given relative node URL path i.e. for building
// case-insensitive lookup tables. Idempotent function. Removes any
// order numbers, as well as leading and trailing slashes.
//
//   /foo/bar/  -> foo/bar
//   foo/02_bar -> foo/bar
func normalizeNodeURL(url string) string {
	var normalized []string

	for _, p := range strings.Split(url, "/") {
		if p == "/" {
			continue
		}
		normalized = append(normalized, removeOrderNumber(p))
	}
	return strings.Join(normalized, "/")
}

// Removes order numbers from path/URL segment, if present.
func removeOrderNumber(segment string) string {
	s := NodeTitleRegexp.FindStringSubmatch(segment)

	if len(s) == 0 {
		return segment
	}
	if len(s) > 2 {
		return s[2]
	}
	return s[1]
}
