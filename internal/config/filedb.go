// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
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
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-yaml/yaml"
	"github.com/icza/dyno"
)

func NewFileDB(path string, project string) (*FileDB, error) {
	db := &FileDB{
		path: path,
		data: &Config{
			Org:     "DSK",
			Project: project,
			Lang:    "en",
			Tags:    make([]*TagConfig, 0),
			Sources: []string{"live"},
			Figma:   &FigmaConfig{},
		},
	}
	if err := db.Open(); err != nil {
		return db, err
	}
	h, err := db.CalculateHash()
	if err != nil {
		return db, err
	}
	db.hash = h
	return db, db.Load()
}

type FileDB struct {
	sync.RWMutex

	// Absolute path to the configuration file.
	path string

	// hash over the file's contents.
	hash string

	// file handle to the configuration file.
	file *os.File

	data *Config
}

func (db *FileDB) Open() error {
	db.Lock()
	defer db.Unlock()

	f, err := os.Open(db.path)
	db.file = f
	return err
}

func (db *FileDB) Close() error {
	db.Lock()
	defer db.Unlock()

	return db.file.Close()
}

// IsGone checks if the underlying file still exists. This check
// together with IsStale() can be used to decide whether the need to
// initialize a new struct or can reuse the current one, and update
// it from the files contents.
func (db *FileDB) IsGone() bool {
	db.RLock()
	defer db.RUnlock()

	_, err := os.Stat(db.path)
	return os.IsNotExist(err)
}

func (db *FileDB) IsStale() bool {
	db.RLock()
	defer db.RUnlock()

	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		// The file was existent when we initialized the struct, so
		// its safe to assume, when its gone something fundamentally
		// changed.
		return true
	}

	if db.hash == "" {
		return true
	}

	// The file exists, it is safe to hash.
	h, _ := db.CalculateHash() // Safe, only locks for read.

	return db.hash != h
}

// CalculateHash hash returns a sha1 sum of the file contents.
func (db *FileDB) CalculateHash() (string, error) {
	db.RLock()
	defer db.RUnlock()

	h := sha1.New()
	_, err := io.Copy(h, db.file)
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func (db *FileDB) Refresh() error {
	log.Printf("Refreshing configuration database %s...", db.path)
	db.Load()

	h, err := db.CalculateHash()
	if err != nil {
		return err
	}
	db.Lock()
	db.hash = h
	db.Unlock()
	return nil
}

// Load populates the internal data slice by parsing the underlying
// file.
func (db *FileDB) Load() error {
	db.Lock()
	defer db.Unlock()

	contents, err := ioutil.ReadFile(db.path)
	if err != nil {
		return err
	}

	switch filepath.Ext(db.path) {
	case ".json":
		return json.Unmarshal(contents, &db.data)
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(contents, &db.data); err != nil {
			return err
		}
		db.data.Custom = dyno.ConvertMapI2MapS(db.data.Custom)
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", db.path)
	}
}

func (db *FileDB) Data() *Config {
	db.RLock()
	defer db.RUnlock()
	return db.data
}

func (db *FileDB) IsAcceptedSource(name string) bool {
	db.RLock()
	defer db.RUnlock()

	for _, n := range db.data.Sources {
		ok, _ := filepath.Match(n, name)
		if ok {
			return true
		}
	}
	return false
}
