// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/atelierdisko/dsk/internal/bus"
	"github.com/atelierdisko/dsk/internal/config"
	"github.com/atelierdisko/dsk/internal/ddt"
	"github.com/atelierdisko/dsk/internal/httputil"
	"github.com/atelierdisko/dsk/internal/search"
	"github.com/gorilla/websocket"
)

func NewV1(cdb *config.DB, v string, t *ddt.NodeTree, hub *bus.Broker, s *search.Search) *V1 {
	return &V1{
		configDB: cdb,
		version:  v,
		tree:     t,
		search:   s,
		messages: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

type V1 struct {
	configDB *config.DB

	version string

	tree *ddt.NodeTree

	search *search.Search

	// We subscribe to the broker in our messages endpoint.
	messages *bus.Broker

	// Upgrades HTTP requests to WebSocket-requests.
	upgrader websocket.Upgrader
}

type V1Hello struct {
	Hello   string `json:"hello"`
	Org     string `json:"org"`
	Project string `json:"project"`
	Version string `json:"version"`
}

type V1Config struct {
	*config.Config
}

type V1Node struct {
	Hash        string          `json:"hash"`
	URL         string          `json:"url"`
	Parent      *V1RefNode      `json:"parent"`
	Children    []*V1RefNode    `json:"children"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Authors     []*V1NodeAuthor `json:"authors"`
	Modified    int64           `json:"modified"`
	Version     string          `json:"version"`
	Tags        []string        `json:"tags"`
	Custom      interface{}     `json:"custom"`
	Docs        []*V1NodeDoc    `json:"docs"`
	Assets      []*V1NodeAsset  `json:"assets"`
	Crumbs      []*V1RefNode    `json:"crumbs"`
	Related     []*V1RefNode    `json:"related"`
	Prev        *V1RefNode      `json:"prev"`
	Next        *V1RefNode      `json:"next"`

	// Deprecated, to be removed in APIv3, please use Assets:
	Downloads []*V1NodeAsset `json:"downloads"`
}

// V1TreeMode is a light top down representation of a part of the DDT.
type V1TreeNode struct {
	Hash     string        `json:"hash"`
	URL      string        `json:"url"`
	Children []*V1TreeNode `json:"children"`
	Title    string        `json:"title"`
}

// V1NodeRef have no parent and children. References must be looked
// up using the URL to get more information about them.
type V1RefNode struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type V1NodeTree struct {
	Hash  string      `json:"hash"`
	Root  *V1TreeNode `json:"root"`
	Total uint16      `json:"total"`
}

type V1NodeAuthor struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type V1NodeDoc struct {
	Title      string                `json:"title"`
	HTML       string                `json:"html"`
	Raw        string                `json:"raw"`
	Components []*V1NodeDocComponent `json:"components"`
}

type V1NodeDocComponent struct {
	Raw      string `json:"raw"`
	Position int    `json:"position"`
}

type V1NodeAsset struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Modified int64  `json:"modified"`
	Size     int64  `json:"size"`
}

type V1SearchResults struct {
	URLs  []string `json:"urls"`
	Total int      `json:"total"`
	Took  int64    `json:"took"` // nanoseconds
}

type V1Message struct {
	Topic string `json:"topic"`
	Text  string `json:"text"`

	// Deprecated in favor of Topic
	Typ string `json:"type"`
}

func (api V1) MountHTTPHandlers() {
	http.HandleFunc("/api/v1/hello", api.HelloHandler)
	http.HandleFunc("/api/v1/config", api.ConfigHandler)
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
	http.HandleFunc("/api/v1", api.NotFoundHandler)
}

func (api V1) NewHello() *V1Hello {
	c := api.configDB.Data()
	return &V1Hello{"dsk", c.Org, c.Project, api.version}
}

func (api V1) NewNode(n *ddt.Node) (*V1Node, error) {
	hash, err := n.CalculateHash()
	if err != nil {
		return nil, err
	}

	var parent *V1RefNode
	if n.Parent != nil {
		parent = &V1RefNode{n.Parent.URL(), n.Parent.Title()}
	}

	children := make([]*V1RefNode, 0, len(n.Children))
	for _, v := range n.Children {
		children = append(children, &V1RefNode{v.URL(), v.Title()})
	}

	authors := make([]*V1NodeAuthor, 0)
	for _, author := range n.Authors() {
		authors = append(authors, &V1NodeAuthor{author.Email, author.Name})
	}

	// Fall back to file system based retrieval if a repository is not
	// available. Also covers present but uncommitted files.
	var modified int64
	nModified, err := n.Modified()
	if err != nil {
		return nil, err
	}
	if !nModified.IsZero() {
		modified = nModified.Unix()
	}

	nDocs, err := n.Docs()
	docs := make([]*V1NodeDoc, 0, len(nDocs))
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

		nComponents, _ := v.Components()
		components := make([]*V1NodeDocComponent, 0, len(nComponents))
		for _, n := range nComponents {
			components = append(components, &V1NodeDocComponent{
				Raw:      n.Raw,
				Position: n.Position,
			})
		}

		docs = append(docs, &V1NodeDoc{
			Title:      v.Title(),
			HTML:       string(html[:]),
			Raw:        string(raw[:]),
			Components: components,
		})
	}

	nAssets, err := n.Assets()
	assets := make([]*V1NodeAsset, 0, len(nAssets))
	if err != nil {
		return nil, err
	}
	for _, v := range nAssets {
		d, err := api.NewNodeAsset(v)
		if err != nil {
			return nil, err
		}
		assets = append(assets, d)
	}

	nCrumbs := n.Crumbs(api.tree.Get)
	crumbs := make([]*V1RefNode, 0, len(nCrumbs))
	for _, n := range nCrumbs {
		crumbs = append(crumbs, &V1RefNode{
			n.URL(), n.Title(),
		})
	}

	nRelated := n.Related(api.tree.Get)
	related := make([]*V1RefNode, 0, len(nRelated))
	for _, n := range nRelated {
		related = append(related, &V1RefNode{
			n.URL(), n.Title(),
		})
	}

	var prev *V1RefNode
	var next *V1RefNode
	prevNode, nextNode, err := api.tree.NeighborNodes(n)
	if err != nil {
		return nil, err
	}
	if prevNode != nil {
		prev = &V1RefNode{
			prevNode.URL(), prevNode.Title(),
		}
	}
	if nextNode != nil {
		next = &V1RefNode{
			nextNode.URL(), nextNode.Title(),
		}
	}

	return &V1Node{
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
		Assets:      assets,
		Crumbs:      crumbs,
		Related:     related,
		Prev:        prev,
		Next:        next,
		Custom:      n.Custom(),

		// Deprecated, to be removed in APIv3:
		Downloads: assets,
	}, nil
}

func (api V1) NewTreeNode(n *ddt.Node) (*V1TreeNode, error) {
	hash, err := n.CalculateHash()
	if err != nil {
		return nil, err
	}

	children := make([]*V1TreeNode, 0, len(n.Children))
	for _, v := range n.Children {
		n, err := api.NewTreeNode(v)
		if err != nil {
			return nil, err
		}
		children = append(children, n)
	}

	return &V1TreeNode{
		Hash:     fmt.Sprintf("%x", hash),
		URL:      n.URL(),
		Children: children,
		Title:    n.Title(),
	}, nil
}

func (api V1) NewNodeTree(t *ddt.NodeTree) (*V1NodeTree, error) {
	root, err := api.NewTreeNode(t.Root)
	if err != nil {
		return nil, err
	}

	return &V1NodeTree{
		// Tree hash is the same as the root nodes'.
		Hash:  root.Hash,
		Root:  root,
		Total: t.TotalNodes(),
	}, err
}

func (api V1) NewNodeAsset(a *ddt.NodeAsset) (*V1NodeAsset, error) {
	var modified int64
	aModified, err := a.Modified()
	if err != nil {
		return nil, err
	}
	if !aModified.IsZero() {
		modified = aModified.Unix()
	}

	size, err := a.Size()
	if err != nil {
		return nil, err
	}
	return &V1NodeAsset{
		URL:      a.URL,
		Name:     a.Name(),
		Modified: modified,
		Size:     size,
	}, nil
}

func (api V1) NewNodeTreeSearchResults(nodes []*ddt.Node, total int, took time.Duration) *V1SearchResults {
	urls := make([]string, 0, len(nodes))
	for _, n := range nodes {
		urls = append(urls, n.URL())
	}
	return &V1SearchResults{urls, total, took.Nanoseconds()}
}

// Says hello :)
func (api V1) HelloHandler(w http.ResponseWriter, r *http.Request) {
	httputil.NewResponder(w, r, "application/json").OK(api.NewHello())
}

func (api V1) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	httputil.NewResponder(w, r, "application/json").OK(&V1Config{Config: api.configDB.Data()})
}

// WebSocket endpoint for receiving notifications.
func (api *V1) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	defer r.Body.Close()

	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	id, messages := api.messages.Subscribe("*")

	for {
		m, ok := <-messages // Blocks until we have a message.
		if !ok {
			// Channel is now closed.
			break
		}
		am, _ := json.Marshal(&V1Message{
			m.(*bus.Message).Topic,
			m.(*bus.Message).Text,

			// Previously we sent tree-changed and tree-syned
			// typs. These strings can be restored from the new
			// Topic field, as the topics follow similar names.
			strings.Replace(m.(*bus.Message).Topic, ".", "-", 1),
		})

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
func (api V1) TreeHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()
	// Not getting or checking path, as only tree requests are routed here.

	if wr.Cached(api.tree.CalculateHash) {
		return
	}

	atree, err := api.NewNodeTree(api.tree)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.Cache(api.tree.CalculateHash)
	wr.OK(atree)
}

// Returns information about a single ddt.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table
//   /api/v1/tree/DisplayData/Table/Row
func (api V1) NodeHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := httputil.CheckSafePath(path, api.tree.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	ok, n, err := api.tree.Get(path)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchNode, nil)
		return
	}

	if wr.Cached(n.CalculateHash) {
		return
	}

	an, err := api.NewNode(n)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.Cache(n.CalculateHash)
	wr.OK(an)
}

// Returns a node asset.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table/foo.png
//   /api/v1/tree/DisplayData/Table/Row/bar.mp4
//   /api/v1/tree/DataEntry/Components/Button/test.png
//   /api/v1/tree/Button/foo.mp4
func (api V1) NodeAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	path := r.URL.Path[len("/api/v1/tree/"):]

	if err := httputil.CheckSafePath(path, api.tree.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	ok, n, err := api.tree.Get(filepath.Dir(path))
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchNode, nil)
		return
	}

	a, err := n.Asset(filepath.Base(path))
	if err != nil {
		wr.Error(httputil.ErrNoSuchAsset, err)
		return
	}
	http.ServeFile(w, r, a.Path)
}

// Performs a search over the design defintions tree and returns
// results in form of a flat list of URLs of matched nodes.
//
// Handles this URL:
//   /api/v1/search?q={query}
func (api V1) SearchHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	q := r.URL.Query().Get("q")

	results, total, took, err := api.search.LegacyFilterSearch(q)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	wr.OK(api.NewNodeTreeSearchResults(results, total, took))
}

func (api V1) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	wr.Error(httputil.ErrNotFound, nil)
}
