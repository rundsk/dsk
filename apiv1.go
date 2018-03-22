// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// API provides a layer between our internal and external representation
// of node data. It allows to implement a versioned API with a higher
// guarantee of stability.
package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gamegos/jsend"
)

type APIv1 struct {
	// Instance of the design defintions tree.
	tree *NodeTree
}

type APIv1Node struct {
	Hash        string             `json:"hash"`
	URL         string             `json:"url"`
	Parent      *APIv1RefNode      `json:"parent"`
	Children    []*APIv1RefNode    `json:"children"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Authors     []*APIv1NodeAuthor `json:"authors"`
	Modified    int64              `json:"modified"`
	Version     string             `json:"version"`
	Tags        []string           `json:"tags"`
	Docs        []*APIv1NodeDoc    `json:"docs"`
	Downloads   []*APIv1NodeAsset  `json:"downloads"`
	Crumbs      []*APIv1RefNode    `json:"crumbs"`
	Related     []*APIv1RefNode    `json:"related"`
	Prev        *APIv1RefNode      `json:"prev"`
	Next        *APIv1RefNode      `json:"next"`
}

// Used when building trees, omits most fields to lighten
// transport weight. Parent ommited to prevent recursive
// data structure.
type APIv1TreeNode struct {
	Hash     string           `json:"hash"`
	URL      string           `json:"url"`
	Children []*APIv1TreeNode `json:"children"`
	Title    string           `json:"title"`
}

// A node reference has no parent and children. It must be looked
// up to get more information.
type APIv1RefNode struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type APIv1NodeTree struct {
	Hash       string         `json:"hash"`
	Root       *APIv1TreeNode `json:"root"`
	TotalNodes uint16         `json:"total_nodes"`
}

type APIv1NodeAuthor struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type APIv1NodeDoc struct {
	Title string `json:"title"`
	HTML  string `json:"html"`
	Raw   string `json:"raw"`
}

type APIv1NodeAsset struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func (api APIv1) MountHTTPHandlers() {
	http.HandleFunc("/api/v1/tree", api.treeHandler)
	http.HandleFunc("/api/v1/tree/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.nodeAssetHandler(w, r)
		} else {
			api.nodeHandler(w, r)
		}
	})
	http.HandleFunc("/api/v1/search", api.searchHandler)
}

func (api APIv1) NewNode(n *Node) (*APIv1Node, error) {
	hash, err := n.Hash()
	if err != nil {
		return nil, err
	}

	var parent *APIv1RefNode
	if n.Parent != nil {
		parent = &APIv1RefNode{n.Parent.URL(), n.Parent.Title()}
	}

	children := make([]*APIv1RefNode, 0, len(n.Children))
	for _, v := range n.Children {
		children = append(children, &APIv1RefNode{v.URL(), v.Title()})
	}

	authors := make([]*APIv1NodeAuthor, 0)
	for _, author := range n.Authors(api.tree.authors) {
		authors = append(authors, &APIv1NodeAuthor{author.Email, author.Name})
	}

	nModified := n.Modified()
	modified := int64(0)
	if !nModified.IsZero() {
		modified = nModified.Unix()
	}

	nDocs, err := n.Docs()
	docs := make([]*APIv1NodeDoc, 0, len(nDocs))
	if err != nil {
		return nil, err
	}
	for _, v := range nDocs {
		html, err := v.HTML("/api/v1/tree", n.URL(), api.tree.Get)
		if err != nil {
			return nil, err
		}
		raw, err := v.Raw()
		if err != nil {
			return nil, err
		}
		docs = append(docs, &APIv1NodeDoc{
			Title: v.Title(),
			HTML:  string(html[:]),
			Raw:   string(raw[:]),
		})
	}

	nDownloads, err := n.Downloads()
	downloads := make([]*APIv1NodeAsset, 0, len(nDownloads))
	if err != nil {
		return nil, err
	}
	for _, v := range nDownloads {
		downloads = append(downloads, &APIv1NodeAsset{URL: v.URL, Name: v.Name})
	}

	nCrumbs := n.Crumbs(api.tree.Get)
	crumbs := make([]*APIv1RefNode, 0, len(nCrumbs))
	for _, n := range nCrumbs {
		crumbs = append(crumbs, &APIv1RefNode{
			n.URL(), n.Title(),
		})
	}

	nRelated := n.Related(api.tree.Get)
	related := make([]*APIv1RefNode, 0, len(nRelated))
	for _, n := range nRelated {
		related = append(related, &APIv1RefNode{
			n.URL(), n.Title(),
		})
	}

	var prev *APIv1RefNode
	var next *APIv1RefNode
	prevNode, nextNode, err := api.tree.NeighborNodes(n)
	if err != nil {
		return nil, err
	}
	if prevNode != nil {
		prev = &APIv1RefNode{
			prevNode.URL(), prevNode.Title(),
		}
	}
	if nextNode != nil {
		next = &APIv1RefNode{
			nextNode.URL(), nextNode.Title(),
		}
	}

	return &APIv1Node{
		Hash:        fmt.Sprintf("%x", hash),
		URL:         n.URL(),
		Parent:      parent,
		Children:    children,
		Title:       n.Title(),
		Tags:        n.Tags(),
		Description: n.Description(),
		Authors:     authors,
		Modified:    modified,
		Version:     n.Version(),
		Docs:        docs,
		Downloads:   downloads,
		Crumbs:      crumbs,
		Related:     related,
		Prev:        prev,
		Next:        next,
	}, nil
}

func (api APIv1) NewTreeNode(n *Node) (*APIv1TreeNode, error) {
	hash, err := n.Hash()
	if err != nil {
		return nil, err
	}

	children := make([]*APIv1TreeNode, 0, len(n.Children))
	for _, v := range n.Children {
		n, err := api.NewTreeNode(v)
		if err != nil {
			return nil, err
		}
		children = append(children, n)
	}

	return &APIv1TreeNode{
		Hash:     fmt.Sprintf("%x", hash),
		URL:      n.URL(),
		Children: children,
		Title:    n.Title(),
	}, nil
}

func (api APIv1) NewNodeTree(t *NodeTree) (*APIv1NodeTree, error) {
	root, err := api.NewTreeNode(t.Root)

	// Tree hash is the same as the root nodes'.
	hash, err := api.tree.Root.Hash()
	if err != nil {
		return nil, err
	}

	return &APIv1NodeTree{
		Hash:       fmt.Sprintf("%x", hash),
		Root:       root,
		TotalNodes: t.TotalNodes(),
	}, err
}

func (api APIv1) NewNodeTreeSearchResults(nodes []*Node) []string {
	results := make([]string, 0, len(nodes))
	for _, n := range nodes {
		results = append(results, n.URL())
	}
	return results
}

// Returns all nodes in the design defintions tree, as nested nodes.
//
// Handles this URL:
//   /api/v1/tree
func (api APIv1) treeHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	// Not getting or checking path, as only tree requests are routed here.

	hash, err := api.tree.Root.Hash()
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	etag := fmt.Sprintf("%x", hash)

	if etag == r.Header.Get("If-None-Match") {
		wr.WriteHeader(http.StatusNotModified)
		return
	}

	atree, err := api.NewNodeTree(api.tree)
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	wr.Header().Set("Etag", etag)

	wr.
		Data(atree).
		Status(http.StatusOK).
		Send()
}

// Returns information about a single node.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table
//   /api/v1/tree/DisplayData/Table/Row
func (api APIv1) nodeHandler(w http.ResponseWriter, r *http.Request) {
	wr := jsend.Wrap(w)
	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := checkSafePath(path, api.tree.path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, n, err := api.tree.Get(path)
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	if !ok {
		wr.
			Status(http.StatusNotFound).
			Message(fmt.Sprintf("No node %s in tree", path)).
			Send()
		return
	}

	hash, err := n.Hash()
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	etag := fmt.Sprintf("%x", hash)

	if etag == r.Header.Get("If-None-Match") {
		wr.WriteHeader(http.StatusNotModified)
		return
	}

	an, err := api.NewNode(n)
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}
	wr.Header().Set("Etag", etag)

	wr.
		Data(an).
		Status(http.StatusOK).
		Send()
}

// Returns a node asset.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table/foo.png
//   /api/v1/tree/DisplayData/Table/Row/bar.mp4
//   /api/v1/tree/DataEntry/Components/Button/test.png
//   /api/v1/tree/Button/foo.mp4
func (api APIv1) nodeAssetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := checkSafePath(path, api.tree.path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, n, err := api.tree.Get(filepath.Dir(path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, fmt.Sprintf("No node %s in tree", path), http.StatusNotFound)
		return
	}

	asset, err := n.Asset(filepath.Base(path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, asset.path)
	return
}

// Performs a search over the design defintions tree and returns
// results in form of a flat list of URLs of matched nodes.
//
// Handles this URL:
//   /api/v1/search?q={query}
func (api APIv1) searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	wr := jsend.Wrap(w)

	nodes := api.tree.FuzzySearch(q)
	results := api.NewNodeTreeSearchResults(nodes)
	wr.
		Data(results).
		Status(http.StatusOK).
		Send()
}
