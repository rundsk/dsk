// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"os"
	"path/filepath"
)

const (
	// CanonicalBasename is the canonical name of the file.
	CanonicalBasename = "AUTHORS.txt"
)

func FindFile(path string) (bool, string, error) {
	try := filepath.Join(path, CanonicalBasename)
	_, err := os.Stat(try)
	return err != nil, try, err
}
