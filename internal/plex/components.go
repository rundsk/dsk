// Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plex

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	cssEntryNames = []string{
		"index.css",
		"styles.css",
		"main.css",
	}
)

func NewComponents(path string) (*Components, error) {
	log.Printf("Initializing components from path %s...", path)

	path, err := filepath.Abs(path)
	return &Components{
		FS:   http.Dir(path),
		Path: path,
	}, err
}

type Components struct {
	FS http.FileSystem

	Path string

	JSEntryPoint  string
	CSSEntryPoint string
}

func (cmps *Components) Detect() {
	hasFile := func(path string) bool {
		normalizedPath := filepath.Join(cmps.Path, path)
		if _, err := os.Stat(normalizedPath); err == nil {
			return true
		}
		log.Printf("Failed to load %s components at %s", path, normalizedPath)
		return false
	}
	if hasFile("index.js") {
		cmps.JSEntryPoint = "index.js"
	}
	for _, f := range cssEntryNames {
		if hasFile(f) {
			cmps.CSSEntryPoint = f
			break
		}
	}
}
