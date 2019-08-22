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
	"time"
)

func NewAPIv2(t *NodeTree, hub *MessageBroker, s *Search) *APIv2 {
	return &APIv2{
		v1:     NewAPIv1(t, hub, s),
		tree:   t,
		search: s,
	}
}

type APIv2 struct {
	v1     *APIv1
	tree   *NodeTree
	search *Search
}

// APIv2FullSearchResults differs from APIv2FilterResults in some
// important ways: The results may be paginated. FilterResults always
// contains all found results in form of a list of node URLs.
type APIv2FullSearchResults struct {
	Hits  []*APIv2FullSearchHit `json:"hits"`
	Total int                   `json:"total"`
	Took  int64                 `json:"took"` // nanoseconds
}

type APIv2FullSearchHit struct {
	APIv1RefNode
	Description string   `json:"description"`
	Fragments   []string `json:"fragments"`
}

type APIv2FilterResults struct {
	Nodes []*APIv1RefNode `json:"nodes"`
	Total int             `json:"total"`
	Took  int64           `json:"took"` // nanoseconds
}

func (api APIv2) MountHTTPHandlers() {
	http.HandleFunc("/api/v2/hello", api.v1.HelloHandler)
	http.HandleFunc("/api/v2/tree", api.v1.TreeHandler)
	http.HandleFunc("/api/v2/tree/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.v1.NodeAssetHandler(w, r)
		} else {
			api.v1.NodeHandler(w, r)
		}
	})
	http.HandleFunc("/api/v2/filter", api.FilterHandler)
	http.HandleFunc("/api/v2/search", api.SearchHandler)
	http.HandleFunc("/api/v2/messages", api.v1.MessagesHandler)
	http.HandleFunc("/api/v2", api.v1.NotFoundHandler)
}

func (api APIv2) NewNodeTreeSearchResults(hs []*FullSearchHit, total int, took time.Duration) *APIv2FullSearchResults {
	hits := make([]*APIv2FullSearchHit, 0, len(hs))

	for _, hit := range hs {
		hits = append(hits, &APIv2FullSearchHit{
			APIv1RefNode: APIv1RefNode{
				hit.Node.URL(),
				hit.Node.Title(),
			},
			Description: hit.Node.Description(),
			Fragments:   hit.Fragments,
		})
	}
	return &APIv2FullSearchResults{hits, total, took.Nanoseconds()}
}

func (api APIv2) NewNodeTreeFilterResults(nodes []*Node, total int, took time.Duration) *APIv2FilterResults {
	ns := make([]*APIv1RefNode, 0, len(nodes))
	for _, n := range nodes {
		ns = append(ns, &APIv1RefNode{n.URL(), n.Title()})
	}
	return &APIv2FilterResults{ns, total, took.Nanoseconds()}
}

// Performs a full broad search over the design defintions tree.
//
// Handles this URL:
//   /api/v2/search?q={query}
func (api APIv2) SearchHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	r.Body.Close()

	q := r.URL.Query().Get("q")

	results, total, took, _, err := api.search.FullSearch(q)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}

	wr.OK(api.NewNodeTreeSearchResults(results, total, took))
}

// Performs a restricted narrow search over the design defintions tree.
//
// Handles these URLs:
//   /api/v2/filter?q={query}
//   /api/v2/filter?q={query}&index=wide
func (api APIv2) FilterHandler(w http.ResponseWriter, r *http.Request) {
	wr := &HTTPResponder{w, r, "application/json"}
	r.Body.Close()

	q := r.URL.Query().Get("q")

	results, total, took, _, err := api.search.FilterSearch(q)
	if err != nil {
		wr.Error(HTTPErr, err)
		return
	}

	wr.OK(api.NewNodeTreeFilterResults(results, total, took))
}
