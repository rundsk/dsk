// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"io/ioutil"
	textTpl "text/template"

	esbuild "github.com/evanw/esbuild/pkg/api"

	"github.com/rs/cors"
	"github.com/rundsk/dsk/internal/bus"
	"github.com/rundsk/dsk/internal/ddt"
	"github.com/rundsk/dsk/internal/httputil"
	"github.com/rundsk/dsk/internal/plex"
	"github.com/rundsk/dsk/internal/search"
)

var (
	//go:embed *.tmpl
	templatesFS embed.FS

	playgroundIndexTemplate *template.Template
)

func init() {
	playgroundIndexTemplate = template.Must(template.ParseFS(templatesFS, "v2_playground_index.html.tmpl"))
}

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
	root := mux.NewRouter()

	tree := root.PathPrefix("/tree").Subrouter()
	node := tree.PathPrefix("/{node:[0-9a-zA-Z-_/]+}").Subrouter()                   // node is one or multiple slugged path elements.
	doc := node.PathPrefix("/_docs/{doc:[0-9a-zA-Z-_.]+}").Subrouter()               // doc is a single slugged path element, a filename.
	playground := doc.PathPrefix("/_playgrounds/{playground:[0-9a-z]+}").Subrouter() // playground is a sha1.

	root.HandleFunc("/hello", api.v1.HelloHandler)
	root.HandleFunc("/config", api.v1.ConfigHandler)
	root.HandleFunc("/sources", api.v1.SourcesHandler)

	tree.HandleFunc("", api.v1.TreeHandler)

	node.HandleFunc("", api.v1.NodeHandler)
	node.HandleFunc("/{asset:.*}", api.v1.NodeAssetHandler) // catch-all

	playground.HandleFunc("/index.html", api.PlaygroundHandler)
	playground.HandleFunc("/{asset:.*}", api.PlaygroundAssetHandler) // catch-all

	root.HandleFunc("/filter", api.FilterHandler)
	root.HandleFunc("/search", api.SearchHandler)
	root.HandleFunc("/messages", api.v1.MessagesHandler)

	root.HandleFunc("/", api.v1.NotFoundHandler) // catch-all

	// An empty slice of origins indicates that CORS shoule be
	// disabled. If we'd pass an empty slice to the CORS middleware
	// it'd be interpreted to allow all origins. We want to be "secure
	// by default".
	if len(api.allowOrigins) == 0 {
		return root
	}
	return cors.New(cors.Options{
		AllowedOrigins:   api.allowOrigins,
		AllowCredentials: true,
	}).Handler(root)
}

func (api V2) NewTreeSearchResults(hs []*search.FullSearchHit, total int, took time.Duration) *V2FullSearchResults {
	hits := make([]*V2FullSearchHit, 0, len(hs))

	for _, hit := range hs {
		hits = append(hits, &V2FullSearchHit{
			V1RefNode: V1RefNode{
				hit.Node.Id(),
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
		ns = append(ns, &V1RefNode{n.Id(), n.URL(), n.Title()})
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

	v := r.URL.Query().Get("v")

	// As the regex for the "node" route param is too greedy (there's no option
	// to make it ungreedy) this results in:
	//   The-Design-Definitions-Tree/Documents/Playground/_docs/Readme
	//
	// ...whereas we want:
	//   The-Design-Definitions-Tree/Documents/Playground
	//
	// TODO: Once more than this handler starts to use the "node" param, this
	//       function will need to be made more generally available to other handlers
	//       in one form or the other. A good form would be to create a middleware
	//       that looks for the param and if found will automatically lookup up the
	//       corresponding node and make it available in the context.
	nodeURL := func() string {
		path, _ := mux.Vars(r)["node"]

		// Consume path elements until one begins with an underscore.
		var parts []string
		for _, p := range strings.Split(path, "/") {
			if strings.HasPrefix(p, "_") {
				break
			}
			parts = append(parts, p)
		}
		return strings.Join(parts, "/")
	}
	docURL, _ := mux.Vars(r)["doc"]
	playgroundId, _ := mux.Vars(r)["playground"]

	s, err := api.sources.MustGet(v)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	ok, n, err := s.Tree.Get(nodeURL())
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchNode, nil)
		return
	}

	ok, doc, err := n.GetDoc(docURL)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchDoc, nil)
		return
	}

	ok, playground, err := doc.GetPlayground(playgroundId)
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}
	if !ok {
		wr.Error(httputil.ErrNoSuchPlayground, nil)
		return
	}
	err = playgroundIndexTemplate.Execute(w, struct {
		Id  string
		Raw string
	}{
		Id:  playground.Id(),
		Raw: playground.RawInner,
	})
	if err != nil {
		wr.Error(httputil.Err, err)
		return
	}

	tmpPlaygoundInstance, err := ioutil.TempFile(os.TempDir(), "*.jsx")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	playgroundSrc := `import React, {useCallback} from 'react'

	export default () => {
		const onClick = useCallback(() => {
			alert('Oh yeah')
		}, [])

		return <button onClick={onClick}>It's all coming together</button>
	}
	`

	// Remember to clean up the file afterwards
	defer os.Remove(tmpPlaygoundInstance.Name())

	// Example writing to the file
	if _, err = tmpPlaygoundInstance.Write([]byte(playgroundSrc)); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	playgroundRuntime, err := os.ReadFile(filepath.Join("frontend", "src", "playground-runtime.jsx"))

	if err != nil {
		log.Fatal("Cannot read playground runtime")
	}

	type pageData struct {
		RuntimeJS  template.JS
		ImportPath template.JSStr
	}

	js, err := textTpl.New("playground-runtime-instance").Parse(`{{define "RuntimeInstance"}}import ThePlaygroundInQuestion from '{{.ImportPath}}';

{{.RuntimeJS | }}
{{end}}
`)

	// HMMM
	playgroundRuntimeTmp, err := ioutil.TempFile(os.TempDir(), "*.jsx")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	// Remember to clean up the file afterwards
	defer os.Remove(playgroundRuntimeTmp.Name())

	var b bytes.Buffer
	js.ExecuteTemplate(&b, "RuntimeInstance", pageData{
		ImportPath: template.JSStr(template.JSEscapeString(tmpPlaygoundInstance.Name())),
		RuntimeJS:  template.JS(playgroundRuntime),
	})

	if _, err = playgroundRuntimeTmp.Write(b.Bytes()); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	result := esbuild.Build(esbuild.BuildOptions{
		EntryPointsAdvanced: []esbuild.EntryPoint{{InputPath: playgroundRuntimeTmp.Name(),
			OutputPath: id}},
		Outdir:     filepath.Join(api.components.Path),
		Bundle:     true,
		Write:      true,
		NodePaths:  []string{"frontend/node_modules"},
		PublicPath: "/api/v2/playgrounds",
		LogLevel:   esbuild.LogLevelDebug,
	})

	// Close the file
	if err := tmpPlaygoundInstance.Close(); err != nil {
		log.Fatal(err)
	}

	if len(result.Errors) > 0 {
		log.Fatal(result.Errors[0].Text)
		wr.Error(httputil.Err, nil)
	}

	html, err := template.New("playground-template").Parse(`{{define "T"}}<html>
		<head>
		  <meta charset="utf-8"><meta>
			<link href="{{.CSSRoot}}" rel="stylesheet"></link>
			<script src="{{.JSRoot}}" type="application/javascript"></script>
		</head>
		<body>
			<div id="root"></div>
		</body>
	</html>{{end}}
`)

	var tpl bytes.Buffer
	if err := html.ExecuteTemplate(&tpl, "T", map[string]string{
		"JSRoot":  filepath.Join("/api/v2/playgrounds", id+".js"),
		"CSSRoot": api.components.CSSEntryPoint,
	}); err != nil {
		wr.Error(httputil.Err, err)
	}

	wr.OK(tpl.Bytes())
}

// Serves a playground's assets.
func (api V2) PlaygroundAssetHandler(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewResponder(w, r, "")
	r.Body.Close()

	assetPath, _ := mux.Vars(r)["asset"]

	if err := httputil.CheckSafePath(assetPath, api.components.Path); err != nil {
		wr.Error(httputil.ErrUnsafePath, err)
		return
	}

	asset, err := api.components.FS.Open(assetPath)
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
