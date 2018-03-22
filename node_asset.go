// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"os"
	"strconv"
)

// A downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	path string

	// The URL, relative to the design defintion tree root.
	URL string

	// The basename of the file, usually for display purposes.
	Name string
}

// As naivily calculating the checksum over the whole file will
// be slow due to the potential large sizes of i.e. video assets,
// we simply take the modified data and time.
func (a NodeAsset) Hash() ([]byte, error) {
	h := sha1.New()
	i, err := os.Stat(a.path)
	u := i.ModTime().Unix()
	return h.Sum([]byte(strconv.FormatInt(u, 10))), err
}
