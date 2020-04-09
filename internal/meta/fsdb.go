// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meta

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewFSDB() *FSDB {
	return &FSDB{}
}

// FSDB will extract information from the underlying file system.
type FSDB struct{}

// Modified will look at the directory's and all files modification
// times, and recursive into each contained directory's to return
// the most recent modified time.
//
// This function has different semantics than the file system's mtime:
// Most file systems change the mtime of the directory when a new file
// or directory is created inside it, the mtime will not change when a
// file has been modified.
//
// It will not descend into directories it considers hidden (their
// name is prefixed by a dot), except when the given directory itself
// is dot-hidden.
func (db *FSDB) Modified(path string) (time.Time, error) {
	var modified time.Time

	err := filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(path) == f.Name()

			if strings.HasPrefix(path, ".") && !isRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if f.ModTime().After(modified) {
			modified = f.ModTime()
		}
		return nil
	})
	if err != nil {
		return modified, fmt.Errorf("failed to walk directory tree %s: %s", path, err)
	}
	return modified, nil
}

func (db *FSDB) Refresh() error {
	return nil
}
