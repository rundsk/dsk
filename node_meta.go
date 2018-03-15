// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Looks for a node configuration file in given directory, parses the
// file and returns a filled NodeMeta struct. If not file is found
// returns an empty NodeMeta.
func NewNodeMetaFromPath(path string) (NodeMeta, error) {
	var meta NodeMeta
	f := filepath.Join(path, ConfigBasename)

	if _, err := os.Stat(f); os.IsNotExist(err) {
		return meta, nil
	}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(content, &meta); err != nil {
		return meta, fmt.Errorf("Failed parsing %s: %s", prettyPath(f), err)
	}
	return meta, nil
}

// Metadata parsed from node configuration.
type NodeMeta struct {
	Authors     []string // Email addresses of node authors.
	Description string
	Keywords    []string
	Related     []string
	Tags        []string
	Version     string // Freeform version string.
}
