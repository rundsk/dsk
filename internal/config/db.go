// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/atelierdisko/dsk/internal/pathutil"
	"github.com/go-yaml/yaml"
	"golang.org/x/text/unicode/norm"
)

func NewDB(path string) (*DB, error) {
	c := &DB{
		path: path,
		data: &Config{
			Org:      "DSK",
			Project:  norm.NFC.String(filepath.Base(filepath.Dir(path))),
			Lang:     "en",
			Tags:     make([]*TagConfig, 0),
			Versions: ".*",
			Figma:    &FigmaConfig{},
		},
	}
	if err := c.Open(); err != nil {
		return c, err
	}
	h, err := c.CalculateHash()
	if err != nil {
		return c, err
	}
	c.hash = h
	return c, c.Load()
}

type DB struct {
	sync.RWMutex

	// Absolute path to the configuration file.
	path string

	// hash over the file's contents.
	hash string

	// file handle to the configuration file.
	file *os.File

	data *Config
}

func (c *DB) Open() error {
	c.Lock()
	defer c.Unlock()

	f, err := os.Open(c.path)
	c.file = f
	return err
}

func (c *DB) Close() error {
	c.Lock()
	defer c.Unlock()

	return c.file.Close()
}

// IsGone checks if the underlying file still exists. This check
// together with IsStale() can be used to decide whether the need to
// initialize a new struct or can reuse the current one, and update
// it from the files contents.
func (c *DB) IsGone() bool {
	c.RLock()
	defer c.RUnlock()

	_, err := os.Stat(c.path)
	return os.IsNotExist(err)
}

func (c *DB) IsStale() bool {
	c.RLock()
	defer c.RUnlock()

	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		// The file was existent when we initialized the struct, so
		// its safe to assume, when its gone something fundamentally
		// changed.
		return true
	}

	if c.hash == "" {
		return true
	}

	// The file exists, it is safe to hash.
	h, _ := c.CalculateHash() // Safe, only locks for read.

	return c.hash != h
}

// CalculateHash hash returns a sha1 sum of the file contents.
func (c *DB) CalculateHash() (string, error) {
	c.RLock()
	defer c.RUnlock()

	h := sha1.New()
	_, err := io.Copy(h, c.file)
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func (c *DB) Refresh() error {
	c.Load()

	h, err := c.CalculateHash()
	if err != nil {
		return err
	}
	c.Lock()
	c.hash = h
	c.Unlock()
	return nil
}

// Load populates the internal data slice by parsing the underlying
// file.
func (c *DB) Load() error {
	c.Lock()
	defer c.Unlock()

	contents, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}

	switch filepath.Ext(c.path) {
	case ".json":
		return json.Unmarshal(contents, &c)
	case ".yaml", ".yml":
		return yaml.Unmarshal(contents, &c)
	default:
		return fmt.Errorf("unsupported format: %s", pathutil.Pretty(c.path))
	}
}

func (c *DB) Data() *Config {
	c.RLock()
	defer c.RUnlock()
	return c.data
}

// IsValidVersion returns true, if the given version is whitelisted by configuration.
func (c *DB) IsValidVersion(version string) (bool, error) {
	c.RLock()
	defer c.RUnlock()
	return regexp.Match(c.data.Versions, []byte(version))
}
