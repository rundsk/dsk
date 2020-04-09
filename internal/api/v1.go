// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/config"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/httputil"
	"github.com/rundsk/dsk/internal/plex"
)

func NewV1(ss *plex.Sources, appVersion string, b *bus.Broker, allowOrigins []string) *V1 {
	return &V1{
		appVersion:   appVersion,
		allowOrigins: allowOrigins,
		broker:       b,
		sources:      ss,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

type V1 struct {
	sources *plex.Sources

	appVersion string

	// allowOrigins is a list of origins a cross-domain request can be
	// executed from. If the special * value is present in the list, all
	// origins will be allowed. An origin may contain a wildcard (*) to
	// replace 0 or more characters (i.e.: http://*.domain.com). Usage of
	// wildcards implies a small performance penality. Only one wildcard
	// can be used per origin. The default value is *.
	allowOrigins []string

	// We subscribe to the broker in our messages endpoint.
	broker *bus.Broker

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

type V1Tree struct {
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

	// Optional, format dependent, fields.
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
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
	Typ string `json:"type,omitempty"`
}

type V1Sources struct {
	Sources []*V1Source `json:"sources"`
}

type V1Source struct {
	Name    string `json:"name"`
	IsReady bool   `json:"is_ready"`
}

// HTTPMux returns a HTTP mux that can be mounted onto a root mux.
func (api V1) HTTPMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", api.HelloHandler)
	mux.HandleFunc("/config", api.ConfigHandler)
	mux.HandleFunc("/sources", api.SourcesHandler)
	mux.HandleFunc("/tree", api.TreeHandler)
	mux.HandleFunc("/tree/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.NodeAssetHandler(w, r)
		} else {
			api.NodeHandler(w, r)
		}
	})
	mux.HandleFunc("/search", api.SearchHandler)
	mux.HandleFunc("/messages", api.MessagesHandler)
	mux.HandleFunc("/", api.NotFoundHandler)

	// An empty slice of origins indicates that CORS shoule be
	// disabled. If we'd pass an empty slice to the CORS middleware
	// it'd be interpreted to allow all origins. We want to be "secure
	// by default".
	if len(api.allowOrigins) == 0 {
		return mux
	}
	return cors.New(cors.Options{
		AllowedOrigins:   api.allowOrigins,
		AllowCredentials: true,
	}).Handler(mux)
}

func (api V1) NewHello(s *plex.Source) (*V1Hello, error) {
	c := s.ConfigDB.Data()

	return &V1Hello{
		Hello:   "dsk",
		Version: api.appVersion,
		Org:     c.Org,
		Project: c.Project,
	}, nil
}

func (api V1) NewConfig(s *plex.Source) (*V1Config, error) {
	return &V1Config{s.ConfigDB.Data()}, nil
}

func (api V1) NewNode(n *ddt.Node, s *plex.Source) (*V1Node, error) {
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
		html, err := v.HTML("/api/v1/tree", n.URL(), s.Tree.Get, s.Name)
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

	nCrumbs := n.Crumbs(s.Tree.Get)
	crumbs := make([]*V1RefNode, 0, len(nCrumbs))
	for _, n := range nCrumbs {
		crumbs = append(crumbs, &V1RefNode{
			n.URL(), n.Title(),
		})
	}

	nRelated := n.Related(s.Tree.Get)
	related := make([]*V1RefNode, 0, len(nRelated))
	for _, n := range nRelated {
		related = append(related, &V1RefNode{
			n.URL(), n.Title(),
		})
	}

	var prev *V1RefNode
	var next *V1RefNode
	prevNode, nextNode, err := s.Tree.NeighborNodes(n)
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

func (api V1) NewTreeNode(n *ddt.Node, s *plex.Source) (*V1TreeNode, error) {
	hash, err := n.CalculateHash()
	if err != nil {
		return nil, err
	}

	children := make([]*V1TreeNode, 0, len(n.Children))
	for _, v := range n.Children {
		n, err := api.NewTreeNode(v, s)
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

func (api V1) NewTree(t *ddt.Tree, s *plex.Source) (*V1Tree, error) {
	root, err := api.NewTreeNode(t.Root, s)
	if err != nil {
		return nil, err
	}

	return &V1Tree{
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

	_, width, height, err := a.Dimensions()
	if err != nil {
		return nil, err
	}

	return &V1NodeAsset{
		URL:      a.URL,
		Name:     a.Name(),
		Modified: modified,
		Size:     size,

		// Optional, these can be empty.
		Width:  width,
		Height: height,
	}, nil
}

func (api V1) NewTreeSearchResults(nodes []*ddt.Node, total int, took time.Duration) *V1SearchResults {
	urls := make([]string, 0, len(nodes))
	for _, n := range nodes {
		urls = append(urls, n.URL())
	}
	return &V1SearchResults{urls, total, took.Nanoseconds()}
}

func (api V1) NewSources(ss *plex.Sources) (*V1Sources, error) {
	vsources := make([]*V1Source, 0)

	names := ss.WhitelistedNames()
	for _, n := range names {
		_, s, _ := ss.Get(n)

		vsources = append(vsources, &V1Source{
			Name:    s.Name,
			IsReady: s.IsComplete(),
		})
	}
	return &V1Sources{vsources}, nil
}

// Says hello :)
//
// Handles these URLs:
//   /api/v1/hello
func (api V1) HelloHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	_, s, err := api.sources.Get("live")
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	pl, err := api.NewHello(s)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.OK(pl)
}

// ConfigHandler responds with a configuration object.
//
// Handles these URLs:
//   /api/v1/config
func (api V1) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	_, s, err := api.sources.Get("live")
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	pl, err := api.NewConfig(s)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.OK(pl)
}

// WebSocket endpoint for receiving notifications.
//
// Handles these URLs:
//   /api/v1/messages
func (api *V1) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	defer r.Body.Close()

	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	id, messages := api.broker.Subscribe("*")

	for {
		m, ok := <-messages // Blocks until we have a message.
		if !ok {
			// Channel is now closed.
			break
		}
		log.Printf("Sending %s to WebSocket subscribers...", m)

		am := &V1Message{
			Topic: m.Topic,
			Text:  m.Text,
		}
		// Deprecated/BC: Previously we sent tree-changed and
		// tree-syned typs. These strings can be restored from the new
		// Topic field, as the topics follow similar names.
		if m.Topic == "live.tree.synced" {
			am.Typ = "tree.synced"
		} else if m.Topic == "live.tree.changed" {
			am.Typ = "tree.changed"
		}
		jam, _ := json.Marshal(am)

		err = conn.WriteMessage(websocket.TextMessage, jam)
		if err != nil {
			// Silently unsubscribe, the client has gone away.
			break
		}
	}
	api.broker.Unsubscribe(id)
	conn.Close()
}

// Returns all nodes in the design defintions tree, as nested nodes.
//
// Handles these URLs:
//   /api/v1/tree
//   /api/v1/tree&v={version}
func (api V1) TreeHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()
	// Not getting or checking path, as only tree requests are routed here.

	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	if wr.Cached(s.Tree.CalculateHash) {
		return
	}

	atree, err := api.NewTree(s.Tree, s)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.Cache(s.Tree.CalculateHash)
	wr.OK(atree)
}

// Returns information about a single ddt.
//
// Handles these kinds of URLs:
//   /api/v1/tree/DisplayData/Table?v={version}
func (api V1) NodeHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	path := r.URL.Path[len("/tree/"):]
	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	if err := httputil.CheckSafePath(path, s.Tree.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	ok, n, err := s.Tree.Get(path)
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

	an, err := api.NewNode(n, s)
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
//   /api/v1/tree/Button/foo.mp4&v={version}
//   /api/v1/tree/Button/colors.json&v={version}
//   /api/v1/tree/Button/colors.yaml&v={version}
func (api V1) NodeAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/octet-stream")
	r.Body.Close()

	path := r.URL.Path[len("/tree/"):]
	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	if err := httputil.CheckSafePath(path, s.Tree.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	ok, n, err := s.Tree.Get(filepath.Dir(path))
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchNode, nil)
		return
	}

	ok, a, err := n.Asset(filepath.Base(path))
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if ok {
		http.ServeFile(w, r, a.Path)
		return
	}

	// When the requested asset was not found under its exact name,
	// maybe we just received a request for format conversion?

	// Reusing var a *NodeAsset (see above).
	for _, name := range ddt.AlternateNames(filepath.Base(path)) {
		ok, a, err = n.Asset(name)
		if err != nil {
			wr.Error(httputil.Err, err)
			return
		}
		if ok {
			break
		}
	}
	if a == nil {
		wr.Error(httputil.ErrNoSuchAsset, err)
		return
	}

	// We cannot serve the file contents as is, instead
	// we serve the converted contents.
	ok, content, err := a.As(filepath.Ext(path))
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchAsset, err)
		return
	}

	// The modified time is taken from the original file, as the
	// conversion will change when the original changes.
	modified, err := a.Modified()
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	http.ServeContent(w, r, filepath.Base(path), modified, content)
}

// Performs a search over the design defintions tree and returns
// results in form of a flat list of URLs of matched nodes.
//
// Handles these URL:
//   /api/v1/search?q={query}
//   /api/v1/search?q={query}&v={version}
func (api V1) SearchHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	q := r.URL.Query().Get("q")
	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	results, total, took, err := s.Search.LegacyFilterSearch(q)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	wr.OK(api.NewTreeSearchResults(results, total, took))
}

// List available DDT sources.
//
// Handles this URL:
//   /api/v1/sources
func (api V1) SourcesHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json")
	r.Body.Close()

	ss, err := api.NewSources(api.sources)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	wr.OK(ss)
}

func (api V1) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	wr.Error(httputil.ErrNotFound, nil)
}
