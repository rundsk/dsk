// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

type DB interface {
	// Data returns the internal configuration data.
	Data() *Config

	// Refresh updates the internal data from its original source.
	Refresh() error

	// Open prepares the struct for use, it allow structs implementing
	// the interface i.e. to retrieve data from its underlying source.
	Open() error

	// Close is the counterpart to Open()
	Close() error

	// IsAcceptedSource returns true, when the given source is
	// whitelisted by configuration. It must support pattern matching
	// with multiple ('*') and single ('?') character matching.
	IsAcceptedSource(string) bool
}
