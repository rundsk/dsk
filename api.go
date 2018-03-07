// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// API provides a layer between our internal and external representation
// of node data. It allows to implement a versioned API with a higher
// guarantee of stability.
package main

import (
	"net/http"
	"path/filepath"

	"github.com/gamegos/jsend"
)

type APIv1 struct {
	// Instance of the design defintions tree.
	tree *NodeTree
}

type APIv1Node struct {
	URL         string            `json:"url"`
	Order       uint64            `json:"order"`
	Children    []*APIv1Node      `json:"children"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Docs        map[string]string `json:"docs"`
	Downloads   []*APIv1NodeAsset `json:"downloads"`
	Crumbs      []*APIv1NodeCrumb `json:"crumbs"`
	IsGhost     bool              `json:"is_ghost"`
}

// Skips some node fields, to lighten transport weight.
type APIv1LightNode struct {
	URL      string            `json:"url"`
	Children []*APIv1LightNode `json:"children"`
	Title    string            `json:"title"`
	IsGhost  bool              `json:"is_ghost"`
}

type APIv1NodeTree struct {
	Root       *APIv1LightNode `json:"root"`
	TotalNodes uint16          `json:"total_nodes"`
}

type APIv1NodeAsset struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type APIv1NodeCrumb struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (api APIv1) MountHTTPHandlers(m Middleware) {
	http.HandleFunc("/api/v1/tree", m(api.treeHandler))
	http.HandleFunc("/api/v1/tree/", m(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.nodeAssetHandler(w, r)
		} else {
			api.nodeHandler(w, r)
		}
	}))
	http.HandleFunc("/api/v1/search", m(api.searchHandler))
}

func (api APIv1) NewNode(n *Node) (*APIv1Node, error) {
	nChildren := n.Children()
	children := make([]*APIv1Node, len(nChildren))
	for k, v := range nChildren {
		n, err := api.NewNode(v)
		if err != nil {
			return nil, err
		}
		children[k] = n
	}

	nDocs, err := n.Docs(filepath.Join("/api/v1/tree", n.URL()))
	docs := make(map[string]string, len(nDocs))
	if err != nil {
		return nil, err
	}
	for k, v := range nDocs {
		docs[k] = string(v[:])
	}

	nDownloads, err := n.Downloads()
	downloads := make([]*APIv1NodeAsset, len(nDownloads))
	if err != nil {
		return nil, err
	}
	for k, v := range nDownloads {
		downloads[k] = &APIv1NodeAsset{URL: v.URL, Name: v.Name}
	}

	nCrumbs := n.Crumbs()
	crumbs := make([]*APIv1NodeCrumb, len(nCrumbs))
	for k, v := range nCrumbs {
		crumbs[k] = &APIv1NodeCrumb{URL: v.URL, Title: v.Title}
	}

	return &APIv1Node{
		URL:         n.URL(),
		Order:       n.Order(),
		Children:    children,
		Title:       n.Title(),
		Keywords:    n.Keywords(),
		Description: n.Description(),
		Docs:        docs,
		Downloads:   downloads,
		Crumbs:      crumbs,
		IsGhost:     n.IsGhost,
	}, nil
}

func (api APIv1) NewLightNode(n *Node) (*APIv1LightNode, error) {
	nChildren := n.Children()
	children := make([]*APIv1LightNode, len(nChildren))
	for k, v := range nChildren {
		n, err := api.NewLightNode(v)
		if err != nil {
			return nil, err
		}
		children[k] = n
	}

	return &APIv1LightNode{
		URL:      n.URL(),
		Children: children,
		Title:    n.Title(),
		IsGhost:  n.IsGhost,
	}, nil
}

func (api APIv1) NewNodeTree(t *NodeTree) (*APIv1NodeTree, error) {
	root, err := api.NewLightNode(t.Root)

	return &APIv1NodeTree{
		Root:       root,
		TotalNodes: t.TotalNodes(),
	}, err
}

// Returns all nodes in the design defintions tree, as nested nodes. If given
// will filter
//
// Handles this URL:
//   /api/v1/tree
//   /api/v1/tree?q={query}
func (api APIv1) treeHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	wr := jsend.Wrap(w)
	// Not getting or checking path here, as only tree requests are routed
	// here.

	if err := api.tree.Sync(); err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}

	var tree *NodeTree
	if q != "" {
		filtered, err := api.tree.Filter(q)
		if err != nil {
			wr.
				Status(http.StatusInternalServerError).
				Message(err.Error()).
				Send()
			return
		}
		tree = filtered
	} else {
		tree = api.tree
	}

	atree, err := api.NewNodeTree(tree)
	if err != nil {
		wr.
			Status(http.StatusInternalServerError).
			Message(err.Error()).
			Send()
		return
	}

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

	n, err := api.tree.GetSynced(path)
	if err != nil {
		wr.
			Status(http.StatusNotFound).
			Message(err.Error()).
			Send()
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

	n, err := api.tree.Get(filepath.Dir(path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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

// Performs a full text search over the design defintions tree and
// returns results.
//
// Handles these kinds of URLs:
//   /api/v1/search?q={query}
func (api APIv1) searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	wr := jsend.Wrap(w)

	results, err := api.tree.Search(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wr.
		Data(results).
		Status(http.StatusOK).
		Send()
}
