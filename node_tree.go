// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

var (
	// Directory basenames matching the pattern are not descending into
	// and interpreted as a node.
	IgnoreNodesRegexp = regexp.MustCompile(`^(x[-_].*|_.*|\..*|node_modules)$`)
)

type NodeTree struct {
	// Ensures the tree is locked, when it is being synced, to
	// prevent reads in the middle of syncs.
	sync.RWMutex

	// The absolute root path of the tree.
	path string

	// Maps node URL paths to nodes, for quick lookup.
	lookup map[string]*Node

	// Ordered slice of un-normalized node URLs.
	ordered []string

	// The root node and entry point to the acutal tree.
	Root *Node `json:"root"`

	// Authors database, if AUTHORS.txt file exists.
	authors *Authors

	// Changes to the directory tree a send here.
	changes chan notify.EventInfo

	// Quit channel, receiving true, when the tree is de-initialized.
	done chan bool
}

// A func to retrieve nodes from the tree, using the node's relative
// URL. When the node cannot be found ok will be false.
type NodeGetter func(url string) (ok bool, n *Node, err error)

// Returns an unsynced tree from path; you must initialize the Tree
// using Sync() and install auto-syncer before using it.
func NewNodeTree(path string) *NodeTree {
	return &NodeTree{path: path}
}

// One-way sync: updates tree from file system. Recursively crawls
// the given root directory, constructing a tree of nodes.
func (t *NodeTree) Sync() error {
	t.Lock()
	defer t.Unlock()

	var nodes []*Node

	err := filepath.Walk(t.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(t.path) == f.Name()

			if IgnoreNodesRegexp.MatchString(f.Name()) && !isRoot {
				log.Printf("Ignoring node: %s", prettyPath(path))
				return filepath.SkipDir
			}
			nodes = append(nodes, NewNode(path, t.path))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", t.path, err)
	}

	// In the second pass we're doing two thing: add the children
	// to the nodes and build up the lookup tables, as we're already
	// iterating the nodes.
	lookup := make(map[string]*Node)
	ordered := make([]string, 0, len(nodes))

	for _, n := range nodes {
		lookup[n.LookupURL()] = n
		ordered = append(ordered, n.UnnormalizedURL())

		for _, sn := range nodes {
			if filepath.Dir(n.path) == sn.path {
				n.Parent = sn
				continue
			}
			if filepath.Dir(sn.path) == n.path {
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

	// Refresh the authors database; file may appear or disappear between
	// syncs.
	authorsFile := filepath.Join(t.path, AuthorsConfigBasename)
	var as *Authors

	if _, err := os.Stat(authorsFile); err == nil {
		as, err = NewAuthorsFromFile(authorsFile)
		if err != nil {
			return err
		}
	} else {
		as = &Authors{}
	}
	t.authors = as

	log.Printf("Synced tree with %d total nodes", len(lookup))
	return nil
}

func (t *NodeTree) StartAutoSync() error {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	t.changes = make(chan notify.EventInfo, 1)
	t.done = make(chan bool, 1)

	if err := notify.Watch(t.path+"/...", t.changes, notify.All); err != nil {
		return err
	}

	go func() {
	Outer:
		for {
			select {
			case ei := <-t.changes:
				p := ei.Path()

				// Do not match directories below tree root. If we
				// are placed inside an ignored dir, anything would
				// always be ignored. Even if the tree root directory
				// is set to be ignored, do not, as the tree has been
				// intentionally loaded from that directory.
				pp := strings.TrimPrefix(p, t.path+"/")

				for pp != "." {
					b := filepath.Base(pp)
					isRoot := b == pp
					pp = filepath.Dir(pp)

					if IgnoreNodesRegexp.MatchString(b) && !isRoot {
						log.Printf("Ignoring change: %s", prettyPath(p))
						continue Outer
					}
				}
				log.Printf("Change detected, re-syncing tree: %s", prettyPath(p))

				if err := t.Sync(); err != nil {
					log.Printf("Auto-sync failed: %s", err)
				}
			case <-t.done:
				log.Print("Auto-sync process is stopping...")
				return
			}
		}
	}()
	return nil
}

func (t *NodeTree) StopAutoSync() {
	t.done <- true
	notify.Stop(t.changes)
}

// Returns the neighboring previous and next nodes for the given
// current node. When current node is the last or first node, the
// behavior is not to wrap around.
//
// Determines the next node following the given current node. This
// may either be the first child of the given node, if there are none
// the sibling node and - walking up the tree - if there is none the
// parents sibling node. The algorithm for determing the previous
// node is analogous.
func (t NodeTree) NeighborNodes(current *Node) (prev *Node, next *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	var ok bool

	key := sort.SearchStrings(t.ordered, current.UnnormalizedURL())

	// SearchString returns the next unused key, if the given string
	// isn't found.
	if key == len(t.ordered) {
		return nil, nil, fmt.Errorf("No node with URL path '%s' in tree", current.URL())
	}

	// Be sure current node isn't the first node.
	if key != 0 {
		ok, prev, err = t.Get(normalizeNodeURL(t.ordered[key-1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}

	// Check if current node isn't the last node.
	if key != len(t.ordered)-1 {
		ok, next, err = t.Get(normalizeNodeURL(t.ordered[key+1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}
	return prev, next, err
}

// Returns the number of total nodes in the tree.
func (t NodeTree) TotalNodes() uint16 {
	t.RLock()
	defer t.RUnlock()

	return uint16(len(t.lookup))
}

// Retrieves a node from the tree, performs a case-insensitive match.
func (t NodeTree) Get(url string) (ok bool, n *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	if n, ok := t.lookup[lookupNodeURL(url)]; ok {
		return ok, n, nil
	}
	return false, &Node{}, nil
}

// Performs a narrow fuzzy search on the node's visible attributes
// (the title) plus tags & keywords and returns the collected results
// as a flat node list.
func (t NodeTree) FuzzySearch(query string) []*Node {
	start := time.Now()

	t.RLock()
	defer t.RUnlock()

	var results []*Node

	matches := func(source string, target string) bool {
		if source == "" {
			return false
		}
		return strings.Contains(strings.ToLower(target), strings.ToLower(source))
	}

Outer:
	for _, n := range t.lookup {
		if matches(query, n.Title()) {
			results = append(results, n)
			continue Outer
		}
		if matches(query, n.Description()) {
			results = append(results, n)
			continue Outer
		}
		for _, v := range n.Tags() {
			if matches(query, v) {
				results = append(results, n)
				continue Outer
			}
		}
		for _, v := range n.Keywords() {
			if matches(query, v) {
				results = append(results, n)
				continue Outer
			}
		}
	}

	log.Printf(
		"Fuzzy searched tree: %s (%d results in %s)",
		query, len(results), time.Since(start),
	)
	return results
}
