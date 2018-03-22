// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// Basenames matching this pattern are considered configuration files.
	NodeMetaRegexp = regexp.MustCompile(`(?i)^(index|meta)\.(json|ya?ml)$`)

	// Basenames matching the pattern will be ignored when searching
	// for downloadable files in the node's directory.
	IgnoreDownloadsRegexp = regexp.MustCompile(`(?i)^(.*\.(js|css|md|markdown|html?|json|ya?ml)|\..*|dsk.*)$`)

	// Basenames matching this pattern are considered documents.
	NodeDocsRegexp = regexp.MustCompile(`(?i)^.*\.(md|markdown|html?)$`)

	// Characters that are ignored when looking up an URL,
	// i.e. "foo/bar baz" and "foo/barbaz" are than equal.
	NodeLookupURLIgnoreChars = regexp.MustCompile(`[\s\-_]+`)

	// Patterns for extracting order number and title from a node's
	// path/URL segment in the form of 06_Foo. As well as for
	// "slugging" the URL/path segment.
	NodePathTitleRegexp        = regexp.MustCompile(`^0?(\d+)[_,-]+(.*)$`)
	NodePathInvalidCharsRegexp = regexp.MustCompile(`[^A-Za-z0-9-_]`)
	NodePathMultipleDashRegexp = regexp.MustCompile(`-+`)
)

// Constructs a new node using its path in the filesystem. Returns a
// node instance even if uncritical errors happened. This is to not
// interrupt tree creation in many cases. Tree creation must fail once
// a bridging node cannot be constructed.
func NewNode(path string, root string) *Node {
	n := &Node{
		root:     root,
		path:     path,
		Children: make([]*Node, 0),
	}

	if err := n.loadMeta(); err != nil {
		log.Print(err)
	}
	return n
}

// Node represents a directory inside the design definitions tree.
type Node struct {
	// Absolute path to the design defintions tree root.
	root string

	// Absolute path to the node's directory.
	path string

	// The parent node. If this is the root node, left unset.
	Parent *Node

	// A list of children nodes.
	Children []*Node

	// Meta data as parsed from the node configuration file.
	meta NodeMeta

	// Cached hash of the node.
	hash []byte
}

// Loads node meta data from the first config file found. Config files
// are optional.
func (n *Node) loadMeta() error {
	files, err := ioutil.ReadDir(n.path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !NodeMetaRegexp.MatchString(f.Name()) {
			continue
		}
		m, err := NewNodeMeta(filepath.Join(n.path, f.Name()))
		if err != nil {
			return err
		}
		n.meta = m
		return nil
	}
	// No node configuration found.
	return nil
}

func (n *Node) Hash() ([]byte, error) {
	if n.hash != nil {
		return n.hash, nil
	}
	h := sha1.New()
	hcom := sha1.New()

	h.Write([]byte(n.path))

	docs, _ := n.Docs()
	for _, v := range docs {
		hv, err := v.Hash()
		if err != nil {
			return nil, err
		}
		h.Write(hv)
	}

	downloads, _ := n.Downloads()
	for _, v := range downloads {
		hv, err := v.Hash()
		if err != nil {
			return nil, err
		}
		h.Write(hv)
	}

	hcom.Write(h.Sum(nil))
	for _, v := range n.Children {
		hv, err := v.Hash()
		if err != nil {
			return nil, err
		}
		hcom.Write(hv)
	}
	n.hash = hcom.Sum(nil)
	return n.hash, nil
}

// Returns the normalized URL path fragment, that can be used to
// address this node i.e Input/Password.
func (n Node) URL() string {
	if n.root == n.path {
		return ""
	}
	return normalizeNodeURL(strings.TrimPrefix(n.path, n.root+"/"))
}

// Returns the unnormalized URL path fragment.
func (n Node) UnnormalizedURL() string {
	if n.root == n.path {
		return ""
	}
	return strings.TrimPrefix(n.path, n.root+"/")
}

// Returns the normalized and lower cased lookup URL for this node.
func (n Node) LookupURL() string {
	return NodeLookupURLIgnoreChars.ReplaceAllString(
		strings.ToLower(n.URL()),
		"",
	)
}

// An order number, as a hint for outside sorting mechanisms.
func (n Node) Order() uint64 {
	return orderNumber(filepath.Base(n.path))
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
func (n Node) Related(get NodeGetter) []*Node {
	nodes := make([]*Node, 0, len(n.meta.Related))

	for _, r := range n.meta.Related {
		ok, node, err := get(r)
		if err != nil {
			log.Printf("Skipping related in %s: %s", n.URL(), err)
			continue
		}
		if !ok {
			log.Printf("Skipping related in %s: %s not found in tree", n.URL(), r)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
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
		ok, author, _ := as.Get(email)
		if !ok {
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
// The provided tree URL prefix will be used to resolve and make
// relative links inside the document absolute. This is usually
// something like: /api/v1/tree
func (n Node) Docs() ([]*NodeDoc, error) {
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
			path: filepath.Join(n.path, f.Name()),
		})
	}
	return docs, nil
}

// Returns a list of crumbs. The last element is the current active
// one. Does not include a root node.
func (n Node) Crumbs(get NodeGetter) []*Node {
	nodes := make([]*Node, 0)

	parts := strings.Split(strings.TrimSuffix(n.URL(), "/"), "/")
	for index, _ := range parts {
		url := strings.Join(parts[:index+1], "/")

		ok, node, err := get(url)
		if err != nil {
			log.Printf("Skipping crumb in %s: %s", n.URL(), err)
			continue
		}
		if !ok {
			log.Printf("Skipping crumb in %s: %s not found in tree", n.URL(), url)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
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

func lookupNodeURL(url string) string {
	return NodeLookupURLIgnoreChars.ReplaceAllString(
		strings.ToLower(normalizeNodeURL(url)),
		"",
	)
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
