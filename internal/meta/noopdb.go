// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meta

import (
	"time"
)

func NewNoopDB() *NoopDB {
	return &NoopDB{}
}

// NoopDB will extract information from the underlying file system.
type NoopDB struct{}

func (db *NoopDB) Modified(path string) (time.Time, error) {
	return time.Time{}, nil
}

func (db *NoopDB) Refresh() error {
	return nil
}
