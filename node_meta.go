// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Parses given node configuration file into a NodeMeta.
func NewNodeMeta(file string) (NodeMeta, error) {
	var m NodeMeta

	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(d.path) {
	case ".json":
		if err := json.Unmarshal(content, &m); err != nil {
			return m, fmt.Errorf("Failed parsing %s: %s", prettyPath(file), err)
		}
		return m, nil
	}
	return nil, fmt.Errorf("Config not in a supported format: %s", d.path)
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
