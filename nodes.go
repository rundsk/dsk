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
	Root *Node `json:"root"`
}

// Recursively crawls the given root directory, constructing a flat list
// of nodes.
func NewNodeTreeFromPath(root string) (*NodeTree, error) {
	nodes := make(map[string]*Node)

	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			n, nErr := NewNodeFromPath(path, root)
			if nErr != nil {
				red := color.New(color.FgRed).SprintFunc()
				log.Printf("skipping node: %s", red(nErr))
				return nil
			}
			nodes[path] = n
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree %s: %s", root, err)
	}

	for _, n := range nodes {
		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
		}
	}
	return &NodeTree{nodes[root]}, nil
}
