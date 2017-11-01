// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

type Node struct {
	path     string
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Parent   *Node   `json:"parent"`
	Children []*Node `json:"children"`
}

type NodeMeta struct {
	Foo  string    `json:"foo"`
	Demo []PropSet `json:"demo"`
}

type PropSet interface{}

func NewNodeFromPath(path string, root string) *Node {
	var url string
	if path == root {
		url = "/"
	} else {
		url = strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), "/")
	}
	return &Node{
		path:  path,
		URL:   url,
		Title: filepath.Base(path),
	}
}

func (n Node) Crumbs() map[string]string {
	var crumbs map[string]string // maps url to title
	//
	//	parts := strings.Split(n.url, "/")
	//
	//	for _, p := range parts {
	//		title := p
	//		url :=
	//	}
	return crumbs
}

// Returns CSS for the node.
func (n Node) CSS() ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(n.path, "component.css"))
}

// Returns JS for the node.
func (n Node) JS() ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(n.path, "component.js"))
}

// Reads index.json file when present and returns values.
func (n Node) Meta() (NodeMeta, error) {
	var meta NodeMeta

	content, err := ioutil.ReadFile(filepath.Join(n.path, "index.json"))
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(content, &meta); err != nil {
		return meta, err
	}
	return meta, nil
}

func (n Node) GetDemos() ([]PropSet, error) {
	meta, err := n.Meta()
	if err != nil {
		return nil, err
	}
	return meta.Demo, nil
}

func (n Node) GetDemo(index int) (PropSet, error) {
	meta, err := n.Meta()
	if err != nil {
		return nil, err
	}
	return meta.Demo[index], nil
}

// Returns documentation in HTML
func (n Node) Documentation() (template.HTML, error) {
	file := filepath.Join(n.path, "readme.md")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "", nil
	}
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return template.HTML(blackfriday.Run(contents)), nil
}

func NewNodeListFromPath(root string) ([]*Node, error) {
	var nodes []*Node

	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			nodes = append(nodes, NewNodeFromPath(path, root))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build node tree in %s: %s", root, err)
	}

	for _, n := range nodes {
		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
			//			if filepath.Dir(n.path) == sn.path {
			//				n.Parent = sn
			//			}
		}
	}

	return nodes, nil
}
