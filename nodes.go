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

	"github.com/fatih/color"
)

var (
	// Directory basenames matching the pattern are not descending into
	// and interpreted as a node.
	IgnoreNodesRegexp = regexp.MustCompile(`^(x[-_].*|\..*|node_modules)$`)
)

type NodeTree struct {
	// The absolute root path of the tree.
	path string
	// Maps node URL paths to nodes, for quick lookup.
	lookup map[string]*Node
	// The root node and entry point to the acutal tree.
	Root *Node `json:"root"`
}

func NewNodeTreeFromPath(path string) *NodeTree {
	return &NodeTree{path: path}
}

// One-way sync: updates tree from file system. Recursively crawls
// the given root directory, constructing a tree of nodes. Does not
// support symlinks inside the tree.
func (t *NodeTree) Sync() error {
	yellow := color.New(color.FgYellow).SprintFunc()

	var nodes []*Node

	err := filepath.Walk(t.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			if IgnoreNodesRegexp.MatchString(f.Name()) {
				log.Printf("Ignoring node: %s", yellow(prettyPath(path)))
				return filepath.SkipDir
			}
			n, nErr := NewNodeFromPath(path, t.path)
			if nErr != nil {
				return nErr
			}
			if n.IsGhost {
				log.Printf("Ghosted node: %s", yellow(nErr))
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", t.path, err)
	}

	// In the second pass we're doing two thing: add the children
	// to the nodes and build up a lookup table, as we're already
	// iterating the nodes.
	lookup := make(map[string]*Node)

	for _, n := range nodes {
		lookup[n.NormalizedURL()] = n

		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.AddChild(sn)
			}
		}
	}

	// Swap late, in event of error we keep the previous state.
	t.lookup = lookup
	t.Root = lookup[""]
	log.Printf("Established tree lookup table with %d entries", len(lookup))

	return nil
}

// Returns the number of total nodes in the tree.
func (t NodeTree) TotalNodes() uint16 {
	return uint16(len(t.lookup))
}

// Retrieves a node from the tree.
func (t NodeTree) Get(url string) (*Node, error) {
	if n, ok := t.lookup[normalizeNodeURL(url)]; ok {
		return n, nil
	}
	return &Node{}, fmt.Errorf("No node with URL path '%s' in tree", url)
}

// Retrieves a node from tree and ensures it's synced before.
func (t NodeTree) GetSynced(url string) (*Node, error) {
	if n, ok := t.lookup[normalizeNodeURL(url)]; ok {
		if err := n.Sync(); err != nil {
			return n, err
		}
		return n, nil
	}
	return &Node{}, fmt.Errorf("No node with URL path '%s' in tree", url)
}

// First performs a narrow search on the node's visible attributes (=
// title) plus keywords and returns a new non-sparse tree instance
// selecting only given nodes, their parents and all their children.
//
// Filters out any not selected nodes. Descends into branches first,
// then works its way back up the tree filtering out any nodes, that
// are not selected. For selection conditions see check().
//
// Selecting a leaf node, selects all parents. But not the siblings.
//
//           a*
//
//           b*
//
//      c!   d   e
//
// Selecting a node, always selects all its children.
//
//           a*
//
//           b!
//
//      c*   d*   e*
func (t NodeTree) Filter(query string) (*NodeTree, error) {
	return &NodeTree{}, nil
}

// Performs a full text search on the tree and returns a flat list
// of nodes as results.
//
// TODO: Implement :)
func (t NodeTree) Search(query string) ([]*Node, error) {
	var results []*Node
	return results, nil
}
