// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"sync"
)

func NewDB(path string) (*DB, error) {
	as := &DB{
		path: path,
		data: make([]*Author, 0),
	}
	if err := as.Open(); err != nil {
		return as, err
	}
	h, err := as.CalculateHash()
	if err != nil {
		return as, err
	}
	as.hash = h
	return as, as.Load()
}

// DB uses a file commonly known as AUTHORS.txt and provides an
// interface to lookup information. It is fully synchronized.
type DB struct {
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

func (as *DB) Open() error {
	as.Lock()
	defer as.Unlock()

	f, err := os.Open(as.path)
	as.file = f
	return err
}

func (as *DB) Close() error {
	as.Lock()
	defer as.Unlock()

	return as.file.Close()
}

// IsGone checks if the underlying file still exists. This check
// together with IsStale() can be used to decide whether the need to
// initialize a new struct or can reuse the current one, and update
// it from the files contents.
func (as *DB) IsGone() bool {
	as.RLock()
	defer as.RUnlock()

	_, err := os.Stat(as.path)
	return os.IsNotExist(err)
}

func (as *DB) IsStale() bool {
	as.RLock()
	defer as.RUnlock()

	if _, err := os.Stat(as.path); os.IsNotExist(err) {
		// The file was existent when we initialized the struct, so
		// its safe to assume, when its gone something fundamentally
		// changed.
		return true
	}

	if as.hash == "" {
		return true
	}

	// The file exists, it is safe to hash.
	h, _ := as.CalculateHash() // Safe, only locks for read.

	return as.hash != h
}

// CalculateHash returns a sha1 sum of the file contents.
func (as *DB) CalculateHash() (string, error) {
	as.RLock()
	defer as.RUnlock()

	h := sha1.New()
	_, err := io.Copy(h, as.file)
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func (as *DB) Refresh() error {
	as.Load()

	h, err := as.CalculateHash()
	if err != nil {
		return err
	}
	as.Lock()
	as.hash = h
	as.Unlock()
	return nil
}

// Load populates the internal data slice by parsing the underlying
// file.
func (as *DB) Load() error {
	as.Lock()
	defer as.Unlock()

	parsed, err := parse(as.file)
	if err != nil {
		return err
	}

	as.data = make([]*Author, 0)
	for _, pa := range parsed {
		as.data = append(as.data, &Author{pa.Name, pa.Email})
	}
	return nil
}

// Add single author item to the internal data slice.
func (as *DB) Add(a *Author) {
	as.Lock()
	defer as.Unlock()

	as.data = append(as.data, a)
}

// GetByEmail looks an author up by her/his email.
func (as *DB) GetByEmail(email string) (ok bool, a *Author) {
	as.RLock()
	defer as.RUnlock()

	for _, a := range as.data {
		if a.Email == email {
			return true, a
		}
	}
	return false, &Author{}
}
