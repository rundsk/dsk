// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package frontend

import (
	"log"
	"net/http"
	"path/filepath"

	"os"

	"github.com/rundsk/dsk/internal/httputil"

	"github.com/evanw/esbuild/pkg/api"
)

func NewFrontendFromPath(path string, chroot string) (*Frontend, error) {
	log.Printf("Initializing frontend from path %s...", path)
	path, err := filepath.Abs(path)

	// Link custom front end with symlink or some other tomfoolery.
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"frontend/src/index.js"},
		Outfile:     filepath.Join(path, "index.js"),
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		Loader: map[string]api.Loader{
			".js":  api.LoaderJSX,
			".png": api.LoaderFile,
			".svg": api.LoaderFile,
		},
	})

	// TODO: Better error handling
	if len(result.Errors) > 0 {
		os.Exit(1)
	}

	file, err := os.Create(filepath.Join(path, "index.html"))

	if err != nil {
		return nil, err
	}

	file.WriteString(`<html>
  <head>
		<link href="/index.css" rel="stylesheet" type="text/css"></link>
		<script defer type="application/javascript" src="/index.js"></script>
	</head>
	<body>
	  <div id="root"></div>
	</body>
</html
`)

	// End ESBuild built entry point

	return &Frontend{
		fs:     http.Dir(path),
		chroot: chroot,
	}, err
}

func NewFrontendFromEmbedded(chroot string) *Frontend {
	log.Print("Intializing embedded frontend...")

	return &Frontend{fs: assets, chroot: chroot}
}

type Frontend struct {
	fs http.FileSystem

	// chroot is used to verify that a tree traversal is not attempted.
	chroot string
}

func (f Frontend) HTTPMux() http.Handler {
	mux := http.NewServeMux()

	// Handles frontend root document delivery and frontend assets.
	// The frontend is allowed to use any path except /api. We route
	// everything else into the front controller (index.html).
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) != "" {
			f.AssetHandler(w, r)
		} else {
			f.RootHandler(w, r)
		}
	})

	return mux
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

	if err := httputil.CheckSafePath(path, f.chroot); err != nil {
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
