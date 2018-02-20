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

	"github.com/russross/blackfriday"
)

const (
	ConfigBasename = "index.json"
)

var (
	// Basenames matching the pattern will be ignored when searching
	// for downloadable files in the node's directory.
	IgnoreDownloadsRegexp = regexp.MustCompile(`^.*\.(js|css|md|json)$`)

	// A pattern for extracting order number and title from a title in the form of 06_Foo.
	NodeTitleRegexp = regexp.MustCompile(`^0?(\d+)[_,-]+(.*)$`)
)

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

type NodeMeta struct {
	Description string
	Keywords    []string
}

// A downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	path string
	// The basename of the file, usually for display purposes.
	Name string
	// The URL, relative to the design defintion tree root.
	URL string
}

type NodeCrumb struct {
	Title string
	URL   string
}

// Constructs a new synced node using its path in the filesystem.
// Returns a node instance even if uncritical errors happened. In that
// case the node will be flagged as a "ghost" node.
func NewNodeFromPath(path string, root string) (*Node, error) {
	n := &Node{root: root, path: path}

	m, err := n.parseMeta()
	n.IsGhost = err != nil
	n.meta = m

	return n, nil
}

// One way sync: update node meta data from file system.
func (n *Node) Sync() error {
	m, err := n.parseMeta()
	n.IsGhost = err != nil
	n.meta = m

	return nil
}

// The node's computed title with any ordering numbers stripped off, usually for display purposes.
func (n *Node) Title() string {
	return cleanNodeTitle(n.path)
}

// An order number, as a hint for outside sorting mechanisms.
func (n *Node) Order() uint64 {
	title := filepath.Base(n.path)
	s := NodeTitleRegexp.FindStringSubmatch(title)

	if len(s) > 2 {
		parsed, _ := strconv.ParseUint(s[0], 10, 64)
		return parsed
	}
	return 0
}

// Return the unnormalized/raw URL path fragment, that can be used to
// address this node i.e Input/Password.
func (n Node) URL() string {
	return strings.TrimPrefix(n.path, n.root+"/")
}

// Returns the normalized URL i.e. for bulding case-insentive lookup
// tables. Idempotent function.
func (n Node) NormalizedURL() string {
	return normalizeNodeURL(n.URL())
}

// Returns an alphabetically sorted list of keywords.
func (n Node) Keywords() []string {
	keywords := n.meta.Keywords

	sort.Strings(keywords)
	return keywords
}

// Returns the full description of the node.
func (n Node) Description() string {
	return n.meta.Description
}

// Returns the list of children nodes. May be left empty when node is
// used in a flat list of results, where children information is not
// needed.
func (n *Node) Children() []*Node {
	return n.children
}

func (n *Node) AddChild(cn *Node) {
	n.children = append(n.children, cn)
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
	var results []*NodeAsset

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

// Returns a map of markdown files and their parsed HTML. Keys are
// normlized and use lower cased filenames without the suffix.
// "readme.md" (in any casing) is considered the main document.
func (n Node) Docs() (map[string][]byte, error) {
	docs := make(map[string][]byte)

	matches, err := filepath.Glob(filepath.Join(n.path, "*.md"))
	if err != nil || matches == nil {
		return docs, err
	}
	for _, m := range matches {
		k := strings.TrimSuffix(filepath.Base(m), filepath.Ext(m))

		contents, err := ioutil.ReadFile(m)
		if err != nil {
			return docs, err
		}

		html, err := n.markdownToHTML(contents)
		if err != nil {
			return docs, err
		}
		docs[k] = html
	}
	return docs, nil
}

// Returns a list of crumbs. The last element is the current active
// one. Does not include a root element.
func (n Node) Crumbs() []*NodeCrumb {
	var crumbs []*NodeCrumb

	parts := strings.Split(strings.TrimSuffix(n.URL(), "/"), "/")
	for index, _ := range parts {
		crumbs = append(crumbs, &NodeCrumb{
			Title: cleanNodeTitle(strings.Join(parts[:index+1], "/")),
			URL:   strings.Join(parts[:index+1], "/"),
		})
	}
	return crumbs
}

// Reads node configuration file when present and returns values. When file
// is not present will simply return an empty Meta.
func (n *Node) parseMeta() (NodeMeta, error) {
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
		return meta, fmt.Errorf("Failed parsing %s: %s", prettyPath(f), err)
	}
	return meta, nil
}

// Parses markdown and returns HTML. Absolute URLs are build using the node's URL.
func (n Node) markdownToHTML(contents []byte) ([]byte, error) {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags:          blackfriday.CommonHTMLFlags &^ blackfriday.UseXHTML,
		AbsolutePrefix: n.URL(),
	})
	return blackfriday.Run(
		contents,
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(blackfriday.CommonExtensions&^blackfriday.HeadingIDs),
	), nil
}

// Normalizes give node URL path i.e. for bulding case-insentive
// lookup tables. Idempotent function.
func normalizeNodeURL(url string) string {
	return strings.Trim(strings.ToLower(url), "/")
}

func cleanNodeTitle(path string) string {
	title := filepath.Base(path)
	s := NodeTitleRegexp.FindStringSubmatch(title)

	if len(s) == 0 {
		return title
	}
	if len(s) > 2 {
		return s[2]
	}
	return s[1]
}
