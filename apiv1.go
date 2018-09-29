// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// API provides a layer between our internal and external representation
// of node data. It allows to implement a versioned API with a higher
// guarantee of stability.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

func NewAPIv1(t *NodeTree, hub *MessageBroker, s *Search) *APIv1 {
	return &APIv1{
		tree:     t,
		search:   s,
		messages: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

type APIv1 struct {
	tree *NodeTree

	search *Search

	// We subscribe to the broker in our messages endpoint.
	messages *MessageBroker

	// Upgrades HTTP requests to WebSocket-requests.
	upgrader websocket.Upgrader
}

type APIv1Hello struct {
	Hello   string `json:"hello"`
	Project string `json:"project"`
	Version string `json:"version"`
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

// APIv1TreeMode is a light top down representation of a part of the DDT.
type APIv1TreeNode struct {
	Hash     string           `json:"hash"`
	URL      string           `json:"url"`
	Children []*APIv1TreeNode `json:"children"`
	Title    string           `json:"title"`
}

// APIv1NodeRef have no parent and children. References must be looked
// up using the URL to get more information about them.
type APIv1RefNode struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type APIv1NodeTree struct {
	Hash  string         `json:"hash"`
	Root  *APIv1TreeNode `json:"root"`
	Total uint16         `json:"total"`
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
	URL      string `json:"url"`
	Name     string `json:"name"`
	Modified int64  `json:"modified"`
	Size     int64  `json:"size"`
}

type APIv1SearchResults struct {
	URLs  []string `json:"urls"`
	Total int      `json:"total"`
	Took  int64    `json:"took"` // nanoseconds
}

type APIv1Message struct {
	Typ  string `json:"type"`
	Text string `json:"text"`
}

func (api APIv1) MountHTTPHandlers() {
	http.HandleFunc("/api/v1/hello", api.HelloHandler)
	http.HandleFunc("/api/v1/tree", api.TreeHandler)
	http.HandleFunc("/api/v1/tree/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.NodeAssetHandler(w, r)
		} else {
			api.NodeHandler(w, r)
		}
	})
	http.HandleFunc("/api/v1/search", api.SearchHandler)
	http.HandleFunc("/api/v1/messages", api.MessagesHandler)
}

func (api APIv1) NewHello() *APIv1Hello {
	return &APIv1Hello{"dsk", filepath.Base(api.tree.path), Version}
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
	for _, author := range n.Authors(api.tree.Authors) {
		authors = append(authors, &APIv1NodeAuthor{author.Email, author.Name})
	}

	// Fall back to file system based retrieval if a repository is not
	// available. Also covers present but uncommitted files.
	var modified int64
	var nModified time.Time
	if api.tree.Repository != nil {
		nModified, err = n.ModifiedFromRepository(api.tree.Repository)
		if err != nil && err != ErrNoData {
			return nil, err
		}
	}
	if nModified.IsZero() {
		nModified, err = n.Modified()
		if err != nil {
			return nil, err
		}
	}
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
		d, err := api.NewNodeAsset(v)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, d)
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
	if err != nil {
		return nil, err
	}

	return &APIv1NodeTree{
		// Tree hash is the same as the root nodes'.
		Hash:  root.Hash,
		Root:  root,
		Total: t.TotalNodes(),
	}, err
}

func (api APIv1) NewNodeAsset(a *NodeAsset) (*APIv1NodeAsset, error) {
	// Fall back to file system based retrieval if a repository is not
	// available. Also covers present but uncommitted files.
	var modified int64
	var aModified time.Time
	var err error
	if api.tree.Repository != nil {
		aModified, err = a.ModifiedFromRepository(api.tree.Repository)
		if err != nil && err != ErrNoData {
			return nil, err
		}
	}
	if aModified.IsZero() {
		aModified, err = a.Modified()
		if err != nil {
			return nil, err
		}
	}
	if !aModified.IsZero() {
		modified = aModified.Unix()
	}

	size, err := a.Size()
	if err != nil {
		return nil, err
	}
	return &APIv1NodeAsset{
		URL:      a.URL,
		Name:     a.Name(),
		Modified: modified,
		Size:     size,
	}, nil
}

func (api APIv1) NewNodeTreeSearchResults(nodes []*Node, total int, took time.Duration) *APIv1SearchResults {
	urls := make([]string, 0, len(nodes))
	for _, n := range nodes {
		urls = append(urls, n.URL())
	}
	return &APIv1SearchResults{urls, total, took.Nanoseconds()}
}

// Says hello :)
func (api APIv1) HelloHandler(w http.ResponseWriter, r *http.Request) {
	(&HTTPResponder{w, r, "application/json"}).OK(api.NewHello())
}

// WebSocket endpoint for receiving notifications.
func (api *APIv1) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, ""}

	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	id, messages := api.messages.Subscribe()

	for {
		m, ok := <-messages // Blocks until we have a message.
		if !ok {
			// Channel is now closed.
			break
		}
		am, _ := json.Marshal(&APIv1Message{m.(*Message).typ, m.(*Message).text})

		err = conn.WriteMessage(websocket.TextMessage, am)
		if err != nil {
			// Silently unsubscribe, the client has gone away.
			break
		}
	}
	api.messages.Unsubscribe(id)
	conn.Close()
}

// Returns all nodes in the design defintions tree, as nested nodes.
//
// Handles this URL:
//   /api/v1/tree
func (api APIv1) TreeHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	// Not getting or checking path, as only tree requests are routed here.

	if wr.Cached(api.tree.Hash) {
		return
	}

	atree, err := api.NewNodeTree(api.tree)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	wr.Cache(api.tree.Hash)
	wr.OK(atree)
}

// Returns information about a single node.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table
//   /api/v1/tree/DisplayData/Table/Row
func (api APIv1) NodeHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := checkSafePath(path, api.tree.path); err != nil {
		wr.Error(HTTPErrUnsafePath, err)
		return
	}

	ok, n, err := api.tree.Get(path)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	if !ok {
		wr.Error(HTTPErrNoSuchNode, nil)
		return
	}

	if wr.Cached(n.Hash) {
		return
	}

	an, err := api.NewNode(n)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	wr.Cache(n.Hash)
	wr.OK(an)
}

// Returns a node asset.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table/foo.png
//   /api/v1/tree/DisplayData/Table/Row/bar.mp4
//   /api/v1/tree/DataEntry/Components/Button/test.png
//   /api/v1/tree/Button/foo.mp4
func (api APIv1) NodeAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := checkSafePath(path, api.tree.path); err != nil {
		wr.Error(HTTPErrUnsafePath, err)
		return
	}

	ok, n, err := api.tree.Get(filepath.Dir(path))
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}
	if !ok {
		wr.Error(HTTPErrNoSuchNode, nil)
		return
	}

	a, err := n.Asset(filepath.Base(path))
	if err != nil {
		wr.Error(HTTPErrNoSuchAsset, err)
		return
	}
	http.ServeFile(w, r, a.path)
}

// Performs a search over the design defintions tree and returns
// results in form of a flat list of URLs of matched nodes.
//
// Handles this URL:
//   /api/v1/search?q={query}
func (api APIv1) SearchHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	q := r.URL.Query().Get("q")

	results, total, took, err := api.search.FilterSearch(q, false)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}

	wr.OK(api.NewNodeTreeSearchResults(results, total, took))
}
