// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

func NewStaticDB(project string) *StaticDB {
	return &StaticDB{
		data: &Config{
			Org:     "DSK",
			Project: project,
			Lang:    "en",
			Tags:    make([]*TagConfig, 0),
			Sources: []string{"live"},
			Figma:   &FigmaConfig{},
		},
	}
}

// StaticDB is a read-only configuration database.
type StaticDB struct {
	data *Config
}

func (db *StaticDB) Data() *Config {
	return db.data
}

func (db *StaticDB) Refresh() error {
	return nil
}

func (db *StaticDB) Open() error {
	return nil
}

func (db *StaticDB) Close() error {
	return nil
}

func (db *StaticDB) IsAcceptedSource(name string) bool {
	return name == "live"
}
