// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// Returns dimensions for media where this is possible. "ok" indicates
// if the format was supported.
func (a NodeAsset) Dimensions() (ok bool, w int, h int, err error) {
	switch strings.ToLower(filepath.Ext(a.path)) {
	case ".jpg", ".jpeg", ".png":
		f, err := os.Open(a.path)
		if err != nil {
			return true, 0, 0, err
		}
		image, _, err := image.DecodeConfig(f)
		if err != nil {
			return true, 0, 0, err
		}
		return true, image.Width, image.Height, nil
	default:
		return false, 0, 0, nil
	}
}
