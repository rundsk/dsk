// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meta

import (
	"context"
	"time"

	"github.com/rundsk/dsk/internal/vcs"
)

func NewRepoDB(r *vcs.Repo) *RepoDB {
	return &RepoDB{r}
}

// RepoDB uses the information from a repository to extract better meta data.
type RepoDB struct {
	repo *vcs.Repo
}

func (db *RepoDB) Modified(path string) (time.Time, error) {
	return db.repo.ModifiedWithContext(context.Background(), path)
}

func (db *RepoDB) ModifiedWithContext(ctx context.Context, path string) (time.Time, error) {
	return db.repo.ModifiedWithContext(ctx, path)
}

func (db *RepoDB) Refresh() error {
	return nil
}
