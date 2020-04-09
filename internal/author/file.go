// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
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
	if err != nil {
		if os.IsNotExist(err) {
			return false, try, nil
		}
		return false, try, err
	}
	return true, try, nil
}
