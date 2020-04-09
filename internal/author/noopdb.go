// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

func NewNoopDB() *NoopDB {
	return &NoopDB{}
}

type NoopDB struct{}

func (db *NoopDB) Open() error {
	return nil
}

func (db *NoopDB) Close() error {
	return nil
}

func (db *NoopDB) Refresh() error {
	return nil
}

func (db *NoopDB) GetByEmail(email string) (bool, *Author) {
	return false, &Author{}
}
