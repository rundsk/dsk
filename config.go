// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/go-yaml/yaml"
)

type Config struct {
	// Absolute path to directory to look for configuration files.
	path string

	// The name of the organization that this Design System is for, defaults to "DSK".
	Org string `json:"org,omitempty" yaml:"org,omitempty"`

	// The project name, defaults to the basename of the DDT folder.
	Project string `json:"project,omitempty" yaml:"project,omitempty"`

	// Language, the documents are authored in. Mainly used for indexing
	// the documents, defaults to English ("en").
	Lang string `json:"lang,omitempty" yaml:"lang,omitempty"`

	// A slice of configuration objects for specific tags. Allows you to display certain tags in custom colors.
	Tags []*TagConfig `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type TagConfig struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Color string `json:"color,omitempty" yaml:"color,omitempty"`
}

var (
	// ConfigBasenames are the canonical name of the configuration file.
	ConfigBasenameRegexp = regexp.MustCompile(`(?i)^(dsk|dsk\.(json|ya?ml))$`)
)

func NewConfig(path string) *Config {
	return &Config{
		path:    path,
		Org:     "DSK",
		Project: filepath.Base(path),
		Lang:    "en",
	}
}

func (c *Config) Load() (string, error) {
	file, err := c.detectFile(c.path)
	if err != nil {
		return file, err
	}
	if file == "" {
		return file, nil
	}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return file, err
	}

	switch filepath.Ext(file) {
	case ".json":
		return file, json.Unmarshal(contents, &c)
	case ".yaml", ".yml":
		return file, yaml.Unmarshal(contents, &c)
	default:
		return file, fmt.Errorf("Unsupported format: %s", prettyPath(file))
	}
}

func (c *Config) detectFile(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if ConfigBasenameRegexp.MatchString(f.Name()) {
			return filepath.Join(path, f.Name()), nil
		}
	}
	return "", nil
}
