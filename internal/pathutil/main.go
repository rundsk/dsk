// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"path/filepath"
)

var (
	PrettyRoot string
)

func SetPrettyRoot(path string) {
	PrettyRoot = path
}

// Does not include the tree root directort.
func Pretty(path string) string {
	rel, _ := filepath.Rel(filepath.Dir(PrettyRoot), path)
	return rel
}
