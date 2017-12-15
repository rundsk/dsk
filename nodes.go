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

type NodeTree struct {
	// The absolute root path of the tree.
	path string
	// Maps node paths to nodes, for quick lookup.
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
	var nodes []*Node

	err := filepath.Walk(t.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {

			// Ignore directories
			//	- that start with x_ or x- (^x[-_])
			//	- that start with . (^\.)
			//	- node_modules (node_modules)
			matched, err := regexp.MatchString(`^x[-_]|^\.|node_modules`, f.Name())
			if err != nil {
				return err
			}
			if(matched) {
				red := color.New(color.FgYellow).SprintFunc()
				log.Printf("Ignoring node: %s", red(path));
				return filepath.SkipDir
			}

			n, nErr := NewNodeFromPath(path, root)
			if nErr != nil {
				red := color.New(color.FgRed).SprintFunc()
				log.Printf("Ghosting node: %s", red(nErr))
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", root, err)
	}

	// In the second pass we're doing two thing: add the children
	// to the nodes and build up a lookup table, as we're already
	// iterating the nodes.
	lookup := make(map[string]*Node)

	for _, n := range nodes {
		lookup[n.path] = n

		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
		}
	}

	// Swap late, in event of error we keep the previous state.
	t.lookup = lookup
	t.Root = lookup[root]

	return nil
}

// Returns the number of total nodes in the tree.
func (t NodeTree) TotalNodes() uint16 {
	return uint16(len(t.lookup))
}

// Retrieves a node from the tree. The given path must be relative to
// the root of the tree.
func (t NodeTree) Get(path string) (*Node, error) {
	if n, ok := t.lookup[filepath.Join(t.path, path)]; ok {
		return n, nil
	}
	return &Node{}, fmt.Errorf("no node with path %s in tree", path)
}

// Retrieves a node from tree and syncs it.
func (t NodeTree) GetSynced(path string) (*Node, error) {
	if n, ok := t.lookup[filepath.Join(t.path, path)]; ok {
		if err := n.Sync(); err != nil {
			return n, err
		}
		return n, nil
	}
	return &Node{}, fmt.Errorf("no node with path %s in tree", path)
}
