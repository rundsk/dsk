// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package frontend

import (
	"net/http"
	"path/filepath"

	"github.com/atelierdisko/dsk/internal/httputil"
)

func NewFrontendFromPath(path string, treeBase string) (*Frontend, error) {
	path, err := filepath.Abs(path)

	return &Frontend{
		fs:       http.Dir(path),
		treeBase: treeBase,
	}, err
}

func NewFrontendFromEmbedded(treeBase string) *Frontend {
	return &Frontend{fs: assets, treeBase: treeBase}
}

type Frontend struct {
	fs http.FileSystem

	// treeBase is used to verify that a tree traversal is not attempted.
	treeBase string
}

func (f Frontend) MountHTTPHandlers() {
	// Handles frontend root document delivery and frontend assets.
	// The frontend is allowed to use any path except /api. We route
	// everything else into the front controller (index.html).
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			f.AssetHandler(w, r)
		} else {
			f.RootHandler(w, r)
		}
	})
}

// Serves the frontend's index.html.
//
// Handles these kinds of URLs:
//   /
//   /index.html
//   /* <catch all>
func (f *Frontend) RootHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	path := "index.html"

	// Does not check on path, as we only ever serve a single
	// file from here, and that path is hard-coded.

	asset, err := f.fs.Open(path)
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

// Serves the frontend's assets.
//
// Handles these kinds of URLs:
//   /assets/css/base.css
//   /static/css/main.41064805.css
func (f *Frontend) AssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	path := r.URL.Path[len("/"):]

	if err := httputil.CheckSafePath(path, f.treeBase); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	asset, err := f.fs.Open(path)
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
