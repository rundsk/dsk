// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/httputil"
	"github.com/rundsk/dsk/internal/plex"
	"github.com/rundsk/dsk/internal/search"
)

func NewV2(ss *plex.Sources, cmps *plex.Components, appVersion string, b *bus.Broker, allowOrigins []string) *V2 {
	return &V2{
		v1:           NewV1(ss, appVersion, b, allowOrigins),
		allowOrigins: allowOrigins,
		sources:      ss,
		components:   cmps,
	}
}

type V2 struct {
	v1 *V1

	// allowOrigins is a list of origins a cross-domain request can be
	// executed from. If the special * value is present in the list, all
	// origins will be allowed. An origin may contain a wildcard (*) to
	// replace 0 or more characters (i.e.: http://*.domain.com). Usage of
	// wildcards implies a small performance penality. Only one wildcard
	// can be used per origin. The default value is *.
	allowOrigins []string

	sources *plex.Sources

	components *plex.Components
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

// HTTPMux returns a HTTP mux that can be mounted onto a root mux.
func (api V2) HTTPMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", api.v1.HelloHandler)
	mux.HandleFunc("/config", api.v1.ConfigHandler)
	mux.HandleFunc("/sources", api.v1.SourcesHandler)
	mux.HandleFunc("/tree", api.v1.TreeHandler)
	mux.HandleFunc("/tree/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			api.v1.NodeAssetHandler(w, r)
		} else {
			api.v1.NodeHandler(w, r)
		}
	})
	mux.HandleFunc("/filter", api.FilterHandler)
	mux.HandleFunc("/search", api.SearchHandler)
	mux.HandleFunc("/messages", api.v1.MessagesHandler)
	mux.HandleFunc("/", api.v1.NotFoundHandler)

	mux.HandleFunc("/playgrounds/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != ".html" {
			api.PlaygroundAssetHandler(w, r)
		} else {
			api.PlaygroundHandler(w, r)
		}
	})
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
	wr := httputil.NewResponder(w, r, "application/json")
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
	wr := httputil.NewResponder(w, r, "application/json")
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

func (api V2) PlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "text/html")
	r.Body.Close()

	id := strings.TrimSuffix(r.URL.Path[len("/playgrounds/"):], "/index.html")
	log.Printf("Requesting HTML for Playground with ID %s", id)

	html := []byte("<html><h1>Hello World</h1></html>")
	wr.OK(html)
}

// Serves a playground's assets.
func (api V2) PlaygroundAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	path := r.URL.Path[len("/playgrounds/"):]

	if err := httputil.CheckSafePath(path, api.components.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	asset, err := api.components.FS.Open(path)
	if err != nil {
		wr.Error(httputil.ErrNoSuchAsset, err)
		return
	}
	defer asset.Close()

	info, err := asset.Stat()
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), asset)
}
