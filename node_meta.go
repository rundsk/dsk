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

	"github.com/go-yaml/yaml"
)

// Metadata parsed from node configuration.
type NodeMeta struct {
	path string

	// Email addresses of node authors.
	Authors     []string `json:"authors,omitempty" yaml:"authors,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Related     []string `json:"related,omitempty" yaml:"related,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Freeform version string.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Deprecated:
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
}

func (m *NodeMeta) Create() error {
	var b []byte
	var err error

	switch filepath.Ext(m.path) {
	case ".json":
		b, err = json.Marshal(m)
	case ".yaml", ".yml":
		b, err = yaml.Marshal(m)
	default:
		return fmt.Errorf("Unsupported format: %s", prettyPath(m.path))
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.path, b, 0666)
}

func (m *NodeMeta) Load() error {
	contents, err := ioutil.ReadFile(m.path)
	if err != nil {
		return err
	}

	switch filepath.Ext(m.path) {
	case ".json":
		return json.Unmarshal(contents, &m)
	case ".yaml", ".yml":
		return yaml.Unmarshal(contents, &m)
	default:
		return fmt.Errorf("Unsupported format: %s", prettyPath(m.path))
	}
}
