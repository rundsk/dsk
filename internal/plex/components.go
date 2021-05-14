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
		if _, err := os.Stat(path); err == nil {
			return true
		}
		return false
	}
	if hasFile("index.js") {
		cmps.JSEntryPoint = "index.js"
	}
}
