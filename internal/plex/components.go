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
	"strings"
)

const packageJson = "package.json"

var (
	cssEntryNames = []string{
		"index.css",
		"styles.css",
		"main.css",
	}
)

func NewComponents(path string) (*Components, error) {
	log.Printf("Initializing components from path %s...", path)
	// Allow node module resolution for component paths.
	orgFp := filepath.Dir(path)

	nodePath := path

	// given <path>/@rundsk/example-component-library, this sets NODE_PATH to be <path>
	if strings.HasPrefix(filepath.Base(orgFp), "@") {
		nodePath = filepath.Dir(orgFp)
	}

	path, err := filepath.Abs(path)
	return &Components{
		FS:           http.Dir(path),
		Path:         path,
		JSEntryPoint: nodePath,
	}, err
}

type Components struct {
	FS http.FileSystem

	Path string

	PackageName   string
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

	// We could potentially use https://stackoverflow.com/questions/32037150/style-field-in-package-json#comment73005816_32042285, but other than bits of postcss, I haven't seen this approach used in the wild
	for _, prefix := range []string{"build", "dist"} {
		for _, f := range cssEntryNames {
			curr := filepath.Join(prefix, f)
			if hasFile(curr) {
				log.Printf("Using path %s as CSS entry point", curr)
				cmps.CSSEntryPoint = curr
				return
			}
		}
	}
}
