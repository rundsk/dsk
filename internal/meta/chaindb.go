// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meta

import (
	"context"
	"time"
)

func NewChainDB(db0 *RepoDB, db1 DB) *ChainDB {
	return &ChainDB{db0, db1}
}

// ChainDB chains a RepoDB and a fallback DB, always querying the
// first, and falling back to querying the second. This i.e. allows
// to chain a FSDB and a RepoDB to both handle files that are already
// committed and those that are not, yet.
type ChainDB struct {
	db0 *RepoDB
	db1 DB
}

func (db *ChainDB) Modified(path string) (time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	t, err := db.db0.ModifiedWithContext(ctx, path)

	if err != nil || t.IsZero() {
		return db.db1.Modified(path)
	}
	return t, err
}

func (db *ChainDB) Refresh() error {
	return nil
}
