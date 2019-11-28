// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/httputil"
	"github.com/rundsk/dsk/internal/plex"
	"github.com/rundsk/dsk/internal/search"
)

func NewV2(ss *plex.Sources, appVersion string, b *bus.Broker, allowOrigin string) *V2 {
	return &V2{
		v1:          NewV1(ss, appVersion, b, allowOrigin),
		allowOrigin: allowOrigin,
		sources:     ss,
	}
}

type V2 struct {
	v1 *V1

	// The value of the Access-Control-Allow-Origin HTTP header to set, if empty
	// the header will remain unset. See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	// for valid values.
	allowOrigin string

	sources *plex.Sources
}

// V2FullSearchResults differs from V2FilterResults in some
// important ways: The results may be paginated. FilterResults always
// contains all found results in form of a list of node URLs.
type V2FullSearchResults struct {
	Hits  []*V2FullSearchHit `json:"hits"`
	Total int                `json:"total"`
	Took  int64              `json:"took"` // nanoseconds
}

type V2FullSearchHit struct {
	V1RefNode
	Description string   `json:"description"`
	Fragments   []string `json:"fragments"`
}

type V2FilterResults struct {
	Nodes []*V1RefNode `json:"nodes"`
	Total int          `json:"total"`
	Took  int64        `json:"took"` // nanoseconds
}

func (api V2) MountHTTPHandlers() {
	log.Print("Mounting APIv2 HTTP handlers...")

	http.HandleFunc("/api/v2/hello", api.v1.HelloHandler)
	http.HandleFunc("/api/v2/config", api.v1.ConfigHandler)
	http.HandleFunc("/api/v2/sources", api.v1.SourcesHandler)
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

func (api V2) NewTreeSearchResults(hs []*search.FullSearchHit, total int, took time.Duration) *V2FullSearchResults {
	hits := make([]*V2FullSearchHit, 0, len(hs))

	for _, hit := range hs {
		hits = append(hits, &V2FullSearchHit{
			V1RefNode: V1RefNode{
				hit.Node.URL(),
				hit.Node.Title(),
			},
			Description: hit.Node.Description(),
			Fragments:   hit.Fragments,
		})
	}
	return &V2FullSearchResults{hits, total, took.Nanoseconds()}
}

func (api V2) NewTreeFilterResults(nodes []*ddt.Node, total int, took time.Duration) *V2FilterResults {
	ns := make([]*V1RefNode, 0, len(nodes))
	for _, n := range nodes {
		ns = append(ns, &V1RefNode{n.URL(), n.Title()})
	}
	return &V2FilterResults{ns, total, took.Nanoseconds()}
}

// Performs a full broad search over the design defintions tree.
//
// Handles these URLs:
//   /api/v2/search?q={query}
//   /api/v2/search?q={query}&v={version}
func (api V2) SearchHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json", api.allowOrigin)
	r.Body.Close()

	q := r.URL.Query().Get("q")
	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	results, total, took, _, err := s.Search.FullSearch(q)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	wr.OK(api.NewTreeSearchResults(results, total, took))
}

// Performs a restricted narrow search over the design defintions tree.
//
// Handles these URLs:
//   /api/v2/filter?q={query}
//   /api/v2/filter?q={query}&v={version}
func (api V2) FilterHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "application/json", api.allowOrigin)
	r.Body.Close()

	q := r.URL.Query().Get("q")
	v := r.URL.Query().Get("v")

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	results, total, took, _, err := s.Search.FilterSearch(q)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	wr.OK(api.NewTreeFilterResults(results, total, took))
}
