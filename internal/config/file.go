// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
)

var (
	// BasenameRegexps matches on allowed names of the configuration file.
	BasenameRegexp = regexp.MustCompile(`(?i)^(dsk|dsk\.(json|ya?ml))$`)
)

func FindFile(path string) (bool, string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false, "", err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if BasenameRegexp.MatchString(f.Name()) {
			return true, filepath.Join(path, f.Name()), nil
		}
	}
	return false, "", nil
}
