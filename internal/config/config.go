// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

type Config struct {
	// The name of the organization that this Design System is for, defaults to "DSK".
	Org string `json:"org,omitempty" yaml:"org,omitempty"`

	// The project name, defaults to the basename of the DDT folder.
	Project string `json:"project,omitempty" yaml:"project,omitempty"`

	// Language, the documents are authored in. Mainly used for indexing
	// the documents, defaults to English ("en").
	Lang string `json:"lang,omitempty" yaml:"lang,omitempty"`

	// A slice of configuration objects for specific tags. Allows you to display certain tags in custom colors.
	Tags []*TagConfig `json:"tags,omitempty" yaml:"tags,omitempty"`

	// List of sources or source patterns to whitelist DDT sources
	// that can be selected and switched to, by default just the
	// "live" version is allowed. Multiple versions can be matched
	// using patterns. Patterns may include wildcards ('*'), which
	// match any number of characters or ('?') to match a single
	// character.
	//
	// Internally this is known as "sources", externally to the
	// user as "versions". "Sources" as a term is too abstract and
	// "versions" has many meanings to cover this case.
	Sources []string `json:"versions,omitempty" yaml:"versions,omitempty"`

	// Configuration related to figma.com.
	Figma *FigmaConfig `json:"figma,omitempty" yaml:"figma,omitempty"`

	Custom interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`
}

type TagConfig struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Color string `json:"color,omitempty" yaml:"color,omitempty"`
}

type FigmaConfig struct {
	// A generated figma personal access token, used for accessing the Figma API on users behalf.
	AccessToken string `json:"accessToken,omitempty" yaml:"accessToken,omitempty"`
}
