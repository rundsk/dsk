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

// TODO: Settle on final response fields
//
// APIv2SearchResults has few options about what we add to it.
// We could add the context within which a match was found to have occurred.
// That would potentially require some sort of encoding a match scheme.
//
// For now, I'll elect to just maintain the existing node array that filter uses.
//
// Clarify how this should look like in regards to:
//       - pagination of results
//       - freshness flags
//       - additional stats
//       - information on each results, probably just the node ref-URL is not enough
type APIv2SearchResults struct {
	URLs  []string `json:"urls"`
	Total int      `json:"total"`
	Took  int64    `json:"took"` // nanoseconds
}

// TODO: This will probably mirror SearchResults once that is settled excluding
//       pagination and rich result information?
type APIv2FilterResults struct {
	URLs  []string `json:"urls"`
	Total int      `json:"total"`
	Took  int64    `json:"took"` // nanoseconds
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
}

func (api APIv2) NewNodeTreeSearchResults(nodes []*Node, total int, took time.Duration) *APIv2SearchResults {
	urls := make([]string, 0, len(nodes))
	for _, n := range nodes {
		urls = append(urls, n.URL())
	}
	return &APIv2SearchResults{urls, total, took.Nanoseconds()}
}

func (api APIv2) NewNodeTreeFilterResults(nodes []*Node, total int, took time.Duration) *APIv2FilterResults {
	urls := make([]string, 0, len(nodes))
	for _, n := range nodes {
		urls = append(urls, n.URL())
	}
	return &APIv2FilterResults{urls, total, took.Nanoseconds()}
}

// Performs a full broad search over the design defintions tree.
//
// Handles this URL:
//   /api/v2/search?q={query}
func (api APIv2) SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	(&HTTPResponder{w, r, "application/json"}).OK(
		api.NewNodeTreeSearchResults(
			api.search.FullSearch(q),
		),
	)
}

// Performs a restricted narrow search over the design defintions tree.
//
// Handles this URL:
//   /api/v2/filter?q={query}
func (api APIv2) FilterHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	(&HTTPResponder{w, r, "application/json"}).OK(
		api.NewNodeTreeFilterResults(
			api.search.FilterSearch(q),
		),
	)
}
