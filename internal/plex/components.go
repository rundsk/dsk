// Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plex

import (
	"encoding/json"
	"io/ioutil"
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
		"style.css",
		"main.css",
	}
)

type packageJsonFields struct {
	Name string `json:"name"`
}

// NewComponents sniffs around to make sure that a package.json is found
// at the path provided by env var. If it is, and its name doesn't match
// the path (in node module's resolution), we fudge the path by
// dropping it at a symlinked path instead.
func NewComponents(pathEnvVar string) (*Components, error) {
	log.Printf("Initializing components from path %s...", pathEnvVar)
	nodePath := filepath.Clean(pathEnvVar)

	rawPkgJson, err := ioutil.ReadFile(filepath.Join(nodePath, packageJson))
	if err != nil {
		return nil, err
	}

	var pkgJson packageJsonFields
	json.Unmarshal(rawPkgJson, &pkgJson)

	if err != nil {
		return nil, err
	}

	if filepath.Base(nodePath) != pkgJson.Name {
		dir := filepath.Join("dist", pkgJson.Name)

		nodePath = "dist"
		if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		splitName := strings.Split(pkgJson.Name, "/")

		if len(splitName) > 1 {
			os.MkdirAll(filepath.Dir(dir), os.ModePerm)
		}

		err = os.Symlink(filepath.Clean(pathEnvVar), dir)

		if err != nil {
			return nil, err
		}
	}

	path, err := filepath.Abs(nodePath)
	log.Printf("Using path %s as JS entry point", path)
	return &Components{
		FS: http.Dir(pathEnvVar),
		// This is different because the CSS entrypoint doesn't use ESBuild's package.json lookup and NODE_ENV.
		Path:         filepath.Clean(pathEnvVar),
		JSEntryPoint: path,
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
