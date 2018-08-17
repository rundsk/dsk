// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// IsDownloadable is true when the asset should be available for download.
func (a NodeAsset) IsDownloadable() bool {
	return true
}

func (a NodeAsset) Modified() (time.Time, error) {
	f, err := os.Stat(a.path)
	if err != nil {
		return time.Time{}, err
	}
	return f.ModTime(), nil
}

// Dimensions for asset media when these are possible to detect. "ok"
// indicates if the format was supported.
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
