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

	"github.com/fatih/color"
)

type NodeTree struct {
	path       string
	totalNodes uint16
	Root       *Node `json:"root"`
}

func NewNodeTreeFromPath(path string) *NodeTree {
	return &NodeTree{path, 0, &Node{}}
}

// One-way sync: updates tree from file system.
// Recursively crawls the given root directory, constructing a tree of nodes.
func (t *NodeTree) Sync() error {
	var nodes []*Node

	err := filepath.Walk(t.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			n, nErr := NewNodeFromPath(path, root)
			if nErr != nil {
				red := color.New(color.FgRed).SprintFunc()
				log.Printf("ghosting node: %s", red(nErr))
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree %s: %s", root, err)
	}

	for _, n := range nodes {
		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
		}
	}

	// Keep statistics, it's cheap to do it here.
	t.totalNodes = uint16(len(nodes))
	// Assume root node is the first, found in tree walk above.
	t.Root = nodes[0]

	return nil
}

func (t NodeTree) TotalNodes() uint16 {
	return t.totalNodes
}

// Checks if a node with given path (relative to root) exists in the tree.
func (t NodeTree) HasPath(path string) bool {
	var check func(n *Node) bool

	check = func(n *Node) bool {
		if filepath.Join(t.path, path) == n.path {
			return true
		}
		for _, c := range n.Children {
			if check(c) {
				return true
			}
		}
		return false
	}
	return check(t.Root)
}
