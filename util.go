// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "path/filepath"

func prettyPath(path string) string {
	rel, _ := filepath.Rel(root, path)
	return rel
}

// TODO: Ensures given path is absolute and below root path, if not
// will panic. Used for preventing path traversal.
func mustSafePath(path string) string {
	return path
}
