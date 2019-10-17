// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atelierdisko/dsk/internal/meta"

	"golang.org/x/text/unicode/norm"
)

func NewNodeAsset(path string, URL string, mdb meta.DB) *NodeAsset {
	return &NodeAsset{path, URL, mdb}
}

// A downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	Path string

	// The URL, relative to the design defintion tree root.
	URL string

	metaDB meta.DB
}

// Name is the basename of the file without its order number.
func (a NodeAsset) Name() string {
	return removeOrderNumber(norm.NFC.String(filepath.Base(a.Path)))
}

func (a NodeAsset) Title() string {
	base := norm.NFC.String(filepath.Base(a.Path))
	return removeOrderNumber(strings.TrimSuffix(base, filepath.Ext(base)))
}

func (a NodeAsset) Modified() (time.Time, error) {
	return a.metaDB.Modified(a.Path)
}

// Size returns the file size in bytes.
func (a NodeAsset) Size() (int64, error) {
	f, err := os.Stat(a.Path)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

// Dimensions for asset media when these are possible to detect. "ok"
// indicates if the format was supported.
func (a NodeAsset) Dimensions() (ok bool, w int, h int, err error) {
	switch strings.ToLower(filepath.Ext(a.Path)) {
	case ".jpg", ".jpeg", ".png":
		f, err := os.Open(a.Path)
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
