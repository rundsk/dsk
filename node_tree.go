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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// NewNodeTree construct and partially initializes a NodeTree. Returns
// an unsynced tree from path; you must finalize initialization using
// Sync() or by calling Start().
func NewNodeTree(path string, as *Authors, repo *Repository, w *Watcher, b *MessageBroker) *NodeTree {
	return &NodeTree{
		path:       path,
		Authors:    as,
		Repository: repo,
		watcher:    w,
		broker:     b,
		done:       make(chan bool),
	}
}

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
	Authors *Authors

	// Repository, if the tree is version controlled.
	Repository *Repository

	// Changes to the directory tree are watched here.
	watcher *Watcher

	// A place where we can send filtered messages to.
	broker *MessageBroker

	// Quit channel, receiving true, when the tree is de-initialized.
	done chan bool
}

// NodeGetter retrieves nodes from the tree, using the node's relative
// URL. When the node cannot be found ok will be false.
type NodeGetter func(url string) (ok bool, n *Node, err error)

// NodesGetter retrieves all nodes from the tree.
type NodesGetter func() []*Node

// HashGetter returns a calculated (or cached) hash.
type HashGetter func() ([]byte, error)

func (t *NodeTree) Hash() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()
	return t.Root.Hash()
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
func (t *NodeTree) Sync() error {
	start := time.Now()

	t.Lock()
	defer t.Unlock()

	var nodes []*Node

	yellow := color.New(color.FgYellow)

	err := filepath.Walk(t.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(t.path) == f.Name()

			if strings.HasPrefix(f.Name(), ".") && !isRoot {
				return filepath.SkipDir
			}
			n := NewNode(path, t.path)

			if err := n.Load(); err != nil {
				log.Print(yellow.Sprint(err))
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", t.path, err)
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

	total := len(lookup)
	took := time.Since(start)

	// Refresh the authors database; file may appear or disappear between
	// syncs. New entries might be added.
	if _, err := t.Authors.Load(); err != nil {
		log.Print(yellow.Sprintf("Failed loading authors: %s", err))
	}

	defer t.broker.Accept(NewMessage(
		MessageTypeTreeSynced, fmt.Sprintf("%d node/s in %s", total, took),
	))

	log.Printf("Synced tree with %d total node/s in %s", total, took)
	return nil
}

// StartSyncer installs an auto-syncing process.
func (t *NodeTree) StartSyncer() {
	id, watch := t.watcher.Subscribe()

	go func() {
		for {
			select {
			case p := <-watch:
				t.broker.Accept(NewMessage(
					MessageTypeTreeChanged, prettyPath(p.(string)),
				))
				log.Printf("Re-syncing tree...")

				if err := t.Sync(); err != nil {
					log.Printf("Re-sync failed: %s", err)
				}
			case <-t.done:
				log.Print("Stopping tree auto-syncing (received quit)...")
				t.watcher.Unsubscribe(id)
				return
			}
		}
	}()
}

func (t *NodeTree) StopSyncer() {
	t.done <- true
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
func (t *NodeTree) NeighborNodes(current *Node) (prev *Node, next *Node, err error) {
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
func (t *NodeTree) TotalNodes() uint16 {
	t.RLock()
	defer t.RUnlock()

	return uint16(len(t.lookup))
}

// Retrieves a node from the tree, performs a case-insensitive match.
func (t *NodeTree) Get(url string) (ok bool, n *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	if n, ok := t.lookup[lookupNodeURL(url)]; ok {
		return ok, n, nil
	}
	return false, &Node{}, nil
}

func (t *NodeTree) GetAll() []*Node {
	t.RLock()
	defer t.RUnlock()

	ns := make([]*Node, 0, len(t.lookup))
	for _, n := range t.lookup {
		ns = append(ns, n)
	}
	return ns
}

func (t *NodeTree) GetAuthors() *Authors {
	return t.Authors
}
