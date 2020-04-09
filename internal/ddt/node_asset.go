// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/icza/dyno"
	"github.com/rundsk/dsk/internal/meta"

	"golang.org/x/text/unicode/norm"
)

func NewNodeAsset(path string, URL string, mdb meta.DB) *NodeAsset {
	return &NodeAsset{path, URL, mdb}
}

// An emebeddable or otherwise  downloadable file.
type NodeAsset struct {
	// Absolute path to the file.
	Path string

	// The URL, relative to the design defintion tree root.
	URL string

	metaDB meta.DB
}

// Name is the basename of the file. The canonical name of the asset
// intentionally contains (non-functional) order numbers if the are
// used.
func (a NodeAsset) Name() string {
	return norm.NFC.String(filepath.Base(a.Path))
}

// Title derived from the cleaned-up basename.
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

// As returns the node assets' contents, converted to the type
// indicated by given extensions. The extensions should include the a
// leading ".".
func (a NodeAsset) As(targetExt string) (bool, io.ReadSeeker, error) {
	sourceExt := filepath.Ext(a.Path)

	contents, err := ioutil.ReadFile(a.Path)
	if err != nil {
		return true, nil, err
	}
	var data interface{}

	switch sourceExt {
	case ".json":
		if err := json.Unmarshal(contents, &data); err != nil {
			return true, nil, err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(contents, &data); err != nil {
			return true, nil, err
		}
		// Fixup converted format, otherwise we can't convert to JSON.
		data = dyno.ConvertMapI2MapS(data)
	default:
		return false, nil, nil
	}

	switch targetExt {
	case ".json":
		marshalled, err := json.Marshal(data)
		return true, bytes.NewReader(marshalled), err
	case ".yaml", ".yml":
		marshalled, err := yaml.Marshal(data)
		return true, bytes.NewReader(marshalled), err
	default:
		return false, nil, nil
	}
}

func AlternateNames(name string) []string {
	names := make([]string, 0)

	ext := filepath.Ext(name)
	filename := name[0 : len(name)-len(ext)]

	switch ext {
	case ".json":
		names = append(names, filename+".yaml", filename+".yml")
	case ".yaml", ".yml":
		names = append(names, filename+".json")
	}
	return names
}
