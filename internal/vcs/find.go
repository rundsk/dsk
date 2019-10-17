// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	ErrRepoNotFound = errors.New("no repository found")
)

// FindRepo searches for a repository that the given path is located
// in and returns the paths that hold a .git directory.
func FindRepo(path string) (bool, string, string, error) {
	rroot, err := findRepo(path, false)
	if err != nil && err != ErrRepoNotFound {
		return false, "", "", fmt.Errorf("failed to detect repository: %s", err)
	}
	if err == ErrRepoNotFound {
		return false, "", "", nil
	}
	log.Printf("Detected repository support in: %s", rroot)

	// We have found the main repository but must now check
	// if the path that was provided is actually a submodule.

	rsub, err := findRepo(path, true)
	if err != nil && err != ErrRepoNotFound {
		return false, rroot, "", err
	}
	if err == ErrRepoNotFound {
		// No submodule found use standard support.
		return true, rroot, "", nil
	}
	log.Printf("Using repository submodule in: %s", rsub)
	return true, rroot, rsub, nil
}

// Searches beginning at given path up, until it finds a directory
// containing a ".git" directory. We differentiate between submodules
// having a ".git" file and regular repositories where ".git" is an
// actual directory.
func findRepo(treeRoot string, searchSubmodule bool) (string, error) {
	var path = treeRoot

	for path != "." && path != "/" {
		s, err := os.Stat(filepath.Join(path, ".git"))

		if err == nil {
			if searchSubmodule && s.Mode().IsRegular() {
				// We found a submodule path, as .git is a regular file.
				return path, nil
			} else if !searchSubmodule && s.Mode().IsDir() {
				// We found the path of a realy repository.
				return path, nil
			}
		}
		path, err = filepath.Abs(path + "/..")
		if err != nil {
			return "", err
		}
	}
	return "", ErrRepoNotFound
}
