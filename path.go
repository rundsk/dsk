// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	PrettyPathRoot     string
	TreeRootNotFound   = errors.New("no tree root found")
	RepositoryNotFound = errors.New("no repository found")
)

// Does not include the tree root directort.
func prettyPath(path string) string {
	rel, _ := filepath.Rel(filepath.Dir(PrettyPathRoot), path)
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
	return fmt.Errorf("directory traversal detected, failed check: path %s, root %s", path, root)
}

// Tries to find root directory either by looking at args or the
// current working directory. This function needs the full path to the
// binary as a first argument and optionally an explicitly given path
// as the second argument.
func detectTreeRoot(binary string, given string) (string, error) {
	var here string

	if given != "" {
		here = given
	} else {
		// When no path is given as an argument, take the path to the
		// process itself. This makes sure that when opening the binary from
		// Finder the folder it is stored in is used.
		here = filepath.Dir(binary)
	}
	here, err := filepath.Abs(here)
	if err != nil {
		return here, err
	}
	p, err := filepath.EvalSymlinks(here)
	if err != nil {
		return "", err
	}
	return p, nil
}

// Searches beginning at given path up, until it finds a directory
// containing a ".git" directory. We differentiate between submodules
// having a ".git" file and regular repositories where ".git" is an
// actual directory.
func detectRepository(treeRoot string, searchSubmodule bool) (string, error) {
	var path = treeRoot

	for path != "." {
		s, err := os.Stat(filepath.Join(path, ".git"))

		if err == nil {
			if searchSubmodule && s.Mode().IsRegular() {
				return path, nil
			} else if !searchSubmodule && s.Mode().IsDir() {
				return path, nil
			}
		}
		path, err = filepath.Abs(path + "/..")
		if err != nil {
			return path, err
		}
	}
	return "", RepositoryNotFound
}

// Checks if any of the path segments in the given path, matches regexp.
func anyPathSegmentMatches(path string, r *regexp.Regexp) bool {
	for path != "." {
		b := filepath.Base(path)

		if IgnoreNodesRegexp.MatchString(b) {
			return true
		}
		path = filepath.Dir(path)
	}
	return false
}

// Finds an order number embedded into given path segment and
// returns it. If none is found, returns 0.
func orderNumber(segment string) uint64 {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) > 2 {
		parsed, _ := strconv.ParseUint(s[0], 10, 64)
		return parsed
	}
	return 0
}

// Removes order numbers from path segment, if present.
func removeOrderNumber(segment string) string {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) == 0 {
		return segment
	}
	if len(s) > 2 {
		return s[2]
	}
	return s[1]
}
