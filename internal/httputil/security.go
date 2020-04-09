// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Ensures given path is absolute and below root path, if not will
// return error. Used for preventing path traversal. Accepts absolute and
// relative paths.
//
// Although the Go http stack will resolve all kinds of dotted path
// segments and finally redirect to the non-relative path (i.e. `GET
// ../../etc/shadow` becomes `GET /etc/shadow`), this func is used as
// an additional safety measure. It can also be used on other parts of
// the URL that are not safe by default (i.e. the query string).
func CheckSafePath(path string, root string) error {
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
