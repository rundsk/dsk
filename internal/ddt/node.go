// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

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
	"sync"
	"time"

	"github.com/atelierdisko/dsk/internal/author"
	"github.com/atelierdisko/dsk/internal/config"
	"github.com/atelierdisko/dsk/internal/meta"
	"github.com/mozillazg/go-unidecode"
	"golang.org/x/text/unicode/norm"
)

var (
	// Basenames matching this pattern are considered configuration files.
	NodeMetaRegexp = regexp.MustCompile(`(?i)^(index|meta)\.(json|ya?ml)$`)

	// Basenames matching this pattern are considered documents.
	NodeDocsRegexp = regexp.MustCompile(`(?i)^.*\.(md|markdown|html?|txt)$`)

	// Files that are not considered to be assets in addition to node
	// meta and doc files.
	NodeAssetsIgnoreRegexp = regexp.MustCompile(`(?i)^(dsk|dsk\.(json|ya?ml)|AUTHORS\.txt|empty)$`)

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

// NewNode constructs a new Node using its path in the filesystem and
// initalizing fields. The initialization must finalized by using Load().
func NewNode(
	path string,
	root string,
	cdb config.DB,
	mdb meta.DB,
	adb author.DB,
) *Node {
	return &Node{
		Path:     path,
		root:     root,
		Children: make([]*Node, 0),
		meta:     &NodeMeta{},
		configDB: cdb,
		metaDB:   mdb,
		authorDB: adb,
	}
}

// Node represents a directory inside the design definitions tree.
type Node struct {
	// Ensure node is locked for writes, when updating the node's hash
	// value cache.
	sync.RWMutex

	// Absolute path to the node's directory.
	Path string

	// Absolute path to the design defintions tree root.
	root string

	// The parent ddt. If this is the root node, left unset.
	Parent *Node

	// A list of children nodes.
	Children []*Node

	// Meta data as parsed from the node configuration file.
	meta *NodeMeta

	configDB config.DB

	metaDB meta.DB

	authorDB author.DB

	// hash is the lazily cached hash set, than used by
	// CalculateHash(). The calculation is not super expensive on its
	// own but once the top of node tree branch is queried for its
	// hash value, all children will have to calculate their hashes.
	//
	// There is no stale detection, as we assume the Node is entirely
	// re-initialized when it changes by the NodeTree.
	hash string
}

func (n *Node) Create() error {
	return os.Mkdir(n.Path, 0777)
}

// CreateMeta creates a meta file in the node's directory using the
// given name as the file name. The provided Meta struct, does not
// need to have its path initialized, this is done by this function.
func (n *Node) CreateMeta(name string, meta *NodeMeta) error {
	n.meta = meta
	n.meta.path = filepath.Join(n.Path, name)
	return n.meta.Create()
}

// CreateDoc creates a document in the node's directory.
func (n *Node) CreateDoc(name string, contents []byte) error {
	return ioutil.WriteFile(filepath.Join(n.Path, name), contents, 0666)
}

// Load node meta data from the first config file found and further
// initialize Node.
func (n *Node) Load() error {
	files, err := ioutil.ReadDir(n.Path)
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
		n.meta.path = filepath.Join(n.Path, f.Name())
		if err := n.meta.Load(); err != nil {
			return err
		}
		return nil
	}
	// No node configuration found, but is optional.
	return nil
}

// CalculateHash calculates a good enough hash over all aspects of the
// node, including its children. Excludes parent in calculation, as it
// would cause an infinite loop.
//
// Will cache the once calculated hash, and use the cached on if
// exists. The assumption here is that the node will be entirely
// re-initialized when it changes.
func (n *Node) CalculateHash() (string, error) {
	n.RLock()

	if n.hash != "" {
		defer n.RUnlock()
		return n.hash, nil
	}
	n.RUnlock()

	h := sha1.New()
	hcom := sha1.New()

	// Covers parent name changes, too.
	h.Write([]byte(n.Path))

	// To avoid expensive calculation of asset (may be large videos)
	// and doc hashes over the whole underlying files, we instead use
	// the last modified file time. This also includes any meta data
	// files.
	m, err := n.lastModifiedForIdentity()
	if err != nil {
		return "", err
	}
	h.Write([]byte(strconv.FormatInt(m.Unix(), 10)))

	hcom.Write(h.Sum(nil))
	for _, v := range n.Children {
		hv, err := v.CalculateHash()
		if err != nil {
			return "", err
		}
		hcom.Write([]byte(hv))
	}
	n.Lock()
	defer n.Unlock()
	n.hash = fmt.Sprintf("%x", h.Sum(nil))
	return n.hash, nil
}

// Returns the normalized URL path fragment, that can be used to
// address this node i.e Input/Password.
func (n *Node) URL() string {
	if n.root == n.Path {
		return ""
	}
	return normalizeNodeURL(strings.TrimPrefix(n.Path, n.root+"/"))
}

// Returns the unnormalized URL path fragment.
func (n *Node) UnnormalizedURL() string {
	if n.root == n.Path {
		return ""
	}
	return strings.TrimPrefix(n.Path, n.root+"/")
}

// Returns the normalized and lower cased lookup URL for this ddt.
func (n *Node) LookupURL() string {
	return NodeLookupURLIgnoreChars.ReplaceAllString(
		strings.ToLower(n.URL()),
		"",
	)
}

// An order number, as a hint for outside sorting mechanisms.
func (n *Node) Order() uint64 {
	return orderNumber(filepath.Base(n.Path))
}

// Name is the basename of the file without its order number.
func (n *Node) Name() string {
	return removeOrderNumber(norm.NFC.String(filepath.Base(n.Path)))
}

// The node's computed title with any ordering numbers stripped off,
// usually for display purposes. We normalize the title string to
// make sure all special characters are represented in their composed
// form. Some filesystems store filenames in decomposed form. Using
// these directly in the frontend led to visual inconsistencies. See:
// https://blog.golang.org/normalization
func (n *Node) Title() string {
	if n.root == n.Path {
		return n.configDB.Data().Project
	}
	return removeOrderNumber(norm.NFC.String(filepath.Base(n.Path)))
}

// Returns the full description of the ddt.
func (n *Node) Description() string {
	return n.meta.Description
}

func (n *Node) Custom() interface{} {
	return n.meta.Custom
}

// Returns a list of related nodes.
func (n *Node) Related(get NodeGetter) []*Node {
	nodes := make([]*Node, 0, len(n.meta.Related))

	for _, r := range n.meta.Related {
		ok, node, err := get(r)
		if err != nil {
			log.Printf("Skipping related in %s: %s", n.URL(), err)
			continue
		}
		if !ok {
			log.Printf("Skipping related in %s: '%s' not found in tree", n.URL(), r)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// Returns an alphabetically sorted list of tags.
func (n *Node) Tags() []string {
	if n.meta.Tags == nil {
		return make([]string, 0)
	}
	tags := n.meta.Tags

	sort.Strings(tags)
	return tags
}

// Returns a list of keywords terms. Deprecated, will be removed once
// APIv1 search support is removed.
func (n *Node) Keywords() []string {
	if n.meta.Keywords == nil {
		return make([]string, 0)
	}
	return n.meta.Keywords
}

// Returns a list of node authors; wil use the given authors
// database to augment data with full name if possible.
func (n *Node) Authors() []*author.Author {
	r := make([]*author.Author, 0)

	if n.meta.Authors == nil {
		return r
	}
	for _, email := range n.meta.Authors {
		ok, a := n.authorDB.GetByEmail(email)
		if ok {
			r = append(r, a)
		} else {
			r = append(r, &author.Author{Email: email, Name: ""})
		}
	}
	return r
}

// Modified finds the most recent modified time of this node, including assets and docs.
func (n *Node) Modified() (time.Time, error) {
	n.RLock()
	defer n.RUnlock()
	return n.metaDB.Modified(n.Path)
}

// fastModified is used to caluclate the hash of the node. In contrast
// to Modified() which will look at all subdirectories and their
// files, this implementation looks only for files in the current
// directory.
func (n *Node) lastModifiedForIdentity() (time.Time, error) {
	var modified time.Time

	files, err := ioutil.ReadDir(n.Path)
	if err != nil {
		return modified, err
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if f.ModTime().After(modified) {
			modified = f.ModTime()
		}
	}
	return modified, nil
}

func (n *Node) Version() string {
	return n.meta.Version
}

// Asset given its basename.
func (n *Node) Asset(name string) (*NodeAsset, error) {
	path := filepath.Join(n.Path, name)

	assets, err := n.Assets()

	for _, a := range assets {

		// if the filename with stripped ordernumber matches the given input return the first match
		if removeOrderNumber(a.Name()) == name {

			return NewNodeAsset(
				filepath.Join(n.Path, a.Name()),
				filepath.Join(n.URL(), a.Name()),
				n.metaDB,
			), nil
		}
	}

	f, err := os.Stat(path)
	if os.IsNotExist(err) || err != nil {
		return nil, err
	}
	if f.IsDir() {
		return nil, fmt.Errorf("accessing directory as asset: %s", path)
	}

	return NewNodeAsset(
		filepath.Join(n.Path, f.Name()),
		filepath.Join(n.URL(), f.Name()),
		n.metaDB,
	), nil
}

// Assets are all files inside the node directory excluding system
// files, node documents and meta files.
func (n *Node) Assets() ([]*NodeAsset, error) {
	as := make([]*NodeAsset, 0)

	files, err := ioutil.ReadDir(n.Path)
	if err != nil {
		return as, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if NodeMetaRegexp.MatchString(f.Name()) {
			continue
		}
		if NodeDocsRegexp.MatchString(f.Name()) {
			continue
		}
		if NodeAssetsIgnoreRegexp.MatchString(f.Name()) {
			continue
		}
		as = append(as, NewNodeAsset(
			filepath.Join(n.Path, f.Name()),
			filepath.Join(n.URL(), f.Name()),
			n.metaDB,
		))
	}
	return as, nil
}

// Returns a slice of documents for this ddt.
//
// The provided tree URL prefix will be used to resolve and make
// relative links inside the document absolute. This is usually
// something like: /api/v1/tree
func (n *Node) Docs() ([]*NodeDoc, error) {
	docs := make([]*NodeDoc, 0)

	files, err := ioutil.ReadDir(n.Path)
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
			path: filepath.Join(n.Path, f.Name()),
		})
	}
	return docs, nil
}

// Returns a list of crumbs. The last element is the current active
// one. Does not include a root ddt.
func (n *Node) Crumbs(get NodeGetter) []*Node {
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
			log.Printf("Skipping crumb in %s: '%s' not found in tree", n.URL(), url)
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
//
// Some filesystems store paths in decomposed form. Using these in the URL
// led to URL inconsistencies between different OS. We therefore make sure
// characters are composed. See: https://blog.golang.org/normalization
func normalizeNodeURL(url string) string {
	var normalized []string

	for _, p := range strings.Split(url, "/") {
		if p == "/" {
			continue
		}
		p = norm.NFC.String(p)
		p = removeOrderNumber(p)
		p = unidecode.Unidecode(p)
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
