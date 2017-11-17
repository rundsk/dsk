// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var DirectoryTraversalError = errors.New("directory traversal attempted")

func prettyPath(path string) string {
	rel, _ := filepath.Rel(root, path)
	return rel
}

// Ensures given path is absolute and below root path, if not will
// return error. Used for preventing path traversal. Accepts absolute and
// relative paths.
//
// Although the Go http stack will resolve all kinds of dotted path
// segments and finally redirect to the non-relative path (i.e. `GET
// ../../etc/shadow` becomes `GET /etc/shadow`), this func is used as
// an additional safety measure. It can also be used on other parts of
// the URL that are not safe by default (i.e. the query string).
func checkSafePath(path string, root string) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path = filepath.Clean(path)

	if path == root {
		return nil
	}
	if strings.HasPrefix(path, root) {
		return nil
	}
	log.Printf("directory traversal detected, failed check: path %s, root %s", path, root)
	return DirectoryTraversalError
}

// Tries to find root directory either by looking at args or the
// current working directory.
func detectRoot() (string, error) {
	var here string

	if len(os.Args) == 2 {
		here = os.Args[1]
	} else {
		here = filepath.Dir(os.Args[0])
	}
	here, err := filepath.Abs(here)
	if err != nil {
		return here, err
	}
	return filepath.EvalSymlinks(here)
}

// The name of the template to parse is , i.e. "index.html". The template
// must be located inside the data/views/ directory.
func mustPrepareTemplate(name string) *template.Template {
	t := template.New(name)

	html, err := Asset("data/views/" + name)
	if err != nil {
		log.Fatal(err)
	}
	t, err = t.Parse(string(html[:]))
	if err != nil {
		log.Fatal(err)
	}
	return t
}
