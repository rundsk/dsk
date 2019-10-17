// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

func NewTxtDB(path string) (*TxtDB, error) {
	log.Printf("Initializing text based author database from %s...", path)

	db := &TxtDB{
		path: path,
		data: make([]*Author, 0),
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

// DB uses a file commonly known as AUTHORS.txt and provides an
// interface to lookup information. It is fully synchronized.
type TxtDB struct {
	sync.RWMutex

	// path is an absolute path to the authors file.
	path string

	// file handle to the authors file.
	file *os.File

	// hash over the file's contents.
	hash string

	// Internal slice of authors data.
	data []*Author
}

type Author struct {
	Email string
	Name  string
}

func (db *TxtDB) Open() error {
	db.Lock()
	defer db.Unlock()

	f, err := os.Open(db.path)
	db.file = f
	return err
}

func (db *TxtDB) Close() error {
	db.Lock()
	defer db.Unlock()

	return db.file.Close()
}

// IsGone checks if the underlying file still exists. This check
// together with IsStale() can be used to decide whether the need to
// initialize a new struct or can reuse the current one, and update
// it from the files contents.
func (db *TxtDB) IsGone() bool {
	db.RLock()
	defer db.RUnlock()

	_, err := os.Stat(db.path)
	return os.IsNotExist(err)
}

func (db *TxtDB) IsStale() bool {
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

// CalculateHash returns a sha1 sum of the file contents.
func (db *TxtDB) CalculateHash() (string, error) {
	db.RLock()
	defer db.RUnlock()

	h := sha1.New()
	_, err := io.Copy(h, db.file)
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func (db *TxtDB) Refresh() error {
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
func (db *TxtDB) Load() error {
	db.Lock()
	defer db.Unlock()

	db.file.Seek(0, 0)

	parsed, err := parse(db.file)
	if err != nil {
		return err
	}
	log.Printf("Loading %d author/s...", len(parsed))

	db.data = make([]*Author, 0)
	for _, pa := range parsed {
		db.data = append(db.data, &Author{Name: pa.Name, Email: pa.Email})
	}
	return nil
}

// GetByEmail looks an author up by her/his email.
func (db *TxtDB) GetByEmail(email string) (bool, *Author) {
	db.RLock()
	defer db.RUnlock()

	for _, a := range db.data {
		if a.Email == email {
			return true, a
		}
	}
	return false, &Author{}
}
