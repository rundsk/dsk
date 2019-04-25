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

	"golang.org/x/text/unicode/norm"
)

// A downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	path string

	// The URL, relative to the design defintion tree root.
	URL string
}

// Name is the basename of the file without its order number.
func (a NodeAsset) Name() string {
	return removeOrderNumber(norm.NFC.String(filepath.Base(a.path)))
}

func (a NodeAsset) Title() string {
	base := norm.NFC.String(filepath.Base(a.path))
	return removeOrderNumber(strings.TrimSuffix(base, filepath.Ext(base)))
}

func (a NodeAsset) Modified() (time.Time, error) {
	f, err := os.Stat(a.path)
	if err != nil {
		return time.Time{}, err
	}
	return f.ModTime(), nil
}

// ModifiedFromRepository uses a Repository for calculating the
// modified time. This is trying to provide a better solution for
// situations where the modified date on disk may not reflect the
// actual modification date. This is the case when the DDT was checked
// out from Git during a build process step.
func (a NodeAsset) ModifiedFromRepository(repo *Repository) (time.Time, error) {
	return repo.Modified(a.path)
}

// Size returns the file size in bytes.
func (a NodeAsset) Size() (int64, error) {
	f, err := os.Stat(a.path)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
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
