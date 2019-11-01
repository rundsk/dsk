// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/atelierdisko/dsk/internal/author"
	"github.com/atelierdisko/dsk/internal/bus"
	"github.com/atelierdisko/dsk/internal/config"
	"github.com/atelierdisko/dsk/internal/meta"
)

var (
	ErrTreeRootNotFound = errors.New("no tree root found")
)

// NewTree construct and initializes a Tree.
func NewTree(
	path string,
	cdb config.DB,
	adb author.DB,
	mdb meta.DB,
	b *bus.Broker,
) (*Tree, error) {
	log.Printf("Initializing node tree on %s...", path)

	t := &Tree{
		Path:     path,
		configDB: cdb,
		metaDB:   mdb,
		authorDB: adb,
		broker:   b,
	}
	return t, t.Sync()
}

type Tree struct {
	// Ensures the tree is locked, when it is being synced, to
	// prevent reads in the middle of syncs.
	sync.RWMutex

	// The absolute root path of the tree.
	Path string

	// Maps node URL paths to nodes, for quick lookup.
	lookup map[string]*Node

	// Ordered slice of un-normalized node URLs.
	ordered []string

	// The root node and entry point to the acutal tree.
	Root *Node `json:"root"`

	configDB config.DB

	metaDB meta.DB

	authorDB author.DB

	// A place where we can send filtered messages to.
	broker *bus.Broker
}

// NodeGetter retrieves nodes from the tree, using the node's relative
// URL. When the node cannot be found ok will be false.
type NodeGetter func(url string) (ok bool, n *Node, err error)

// NodesGetter retrieves all nodes from the tree.
type NodesGetter func() []*Node

func (t *Tree) String() string {
	return fmt.Sprintf("node tree (...%s)", t.Path[len(t.Path)-10:])
}

func (t *Tree) CalculateHash() (string, error) {
	t.RLock()
	defer t.RUnlock()
	return t.Root.CalculateHash()
}

// Sync recursively crawls the given root directory, constructing a
// tree of nodes. Will rebuild the entire tree on every sync. This
// makes the algorithm really simple - as we don't need to do branch
// selection - but also slow.
//
// Nodes that are discover but fail to finalize their initialization
// using Node.Load() will not be skipped but kept in tree in
// a semi-initialized way. So that the their children are not
// disconnected and no gaps exist in tree branches.
//
// It will not descend into directories it considers hidden (their
// name is prefixed by a dot), except when the given directory itself
// is dot-hidden.
func (t *Tree) Sync() error {
	log.Printf("Syncing %s...", t)

	t.Lock()
	defer t.Unlock()

	start := time.Now()

	var nodes []*Node

	err := filepath.Walk(t.Path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(t.Path) == f.Name()

			if strings.HasPrefix(f.Name(), ".") && !isRoot {
				return filepath.SkipDir
			}
			n := NewNode(
				path,
				t.Path,
				t.configDB,
				t.metaDB,
				t.authorDB,
			)

			if err := n.Load(); err != nil {
				log.Print(err)
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree %s: %s", t.Path, err)
	}

	// In the second pass we're doing two things: add the children
	// to the nodes and build up the lookup tables, as we're already
	// iterating the nodes.
	lookup := make(map[string]*Node)
	ordered := make([]string, 0, len(nodes))

	for _, n := range nodes {
		lookup[n.LookupURL()] = n
		ordered = append(ordered, n.UnnormalizedURL())

		for _, sn := range nodes {
			if filepath.Dir(n.Path) == sn.Path {
				n.Parent = sn
				continue
			}
			if filepath.Dir(sn.Path) == n.Path {
				n.Children = append(n.Children, sn)
				continue
			}
		}
	}
	sort.Strings(ordered)

	// Swap late, in event of error we keep the previous state.
	t.lookup = lookup
	t.ordered = ordered
	t.Root = lookup[""]

	total := len(lookup)
	took := time.Since(start)

	defer t.broker.Accept("tree.synced", fmt.Sprintf("%d node/s in %s", total, took))

	log.Printf("Synced %s with %d total node/s in %s", t, total, took)
	return nil
}

// Returns the neighboring previous and next nodes for the given
// current ddt. When current node is the last or first node, the
// behavior is not to wrap around.
//
// Determines the next node following the given current ddt. This
// may either be the first child of the given node, if there are none
// the sibling node and - walking up the tree - if there is none the
// parents sibling ddt. The algorithm for determing the previous
// node is analogous.
func (t *Tree) NeighborNodes(current *Node) (prev *Node, next *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	var ok bool

	key := sort.SearchStrings(t.ordered, current.UnnormalizedURL())

	// SearchString returns the next unused key, if the given string
	// isn't found.
	if key == len(t.ordered) {
		return nil, nil, fmt.Errorf("no node with URL path '%s' in %s", current.URL(), t)
	}

	// Be sure current node isn't the first ddt.
	if key != 0 {
		ok, prev, err = t.Get(normalizeNodeURL(t.ordered[key-1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}

	// Check if current node isn't the last ddt.
	if key != len(t.ordered)-1 {
		ok, next, err = t.Get(normalizeNodeURL(t.ordered[key+1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}
	return prev, next, err
}

// Returns the number of total nodes in the tree.
func (t *Tree) TotalNodes() uint16 {
	t.RLock()
	defer t.RUnlock()

	return uint16(len(t.lookup))
}

// Retrieves a node from the tree, performs a case-insensitive match.
func (t *Tree) Get(url string) (ok bool, n *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	if n, ok := t.lookup[lookupNodeURL(url)]; ok {
		return ok, n, nil
	}
	return false, &Node{}, nil
}

// GetAll nodes as a flat slice.
func (t *Tree) GetAll() []*Node {
	t.RLock()
	defer t.RUnlock()

	ns := make([]*Node, 0, len(t.lookup))
	for _, n := range t.lookup {
		ns = append(ns, n)
	}
	return ns
}
