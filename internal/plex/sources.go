// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plex

import (
	"errors"
	"fmt"
	"log"

	"github.com/rundsk/dsk/internal/config"
)

func NewSources(cdb config.DB) (*Sources, error) {
	log.Print("Initializing sources...")

	ss := &Sources{
		Teardown: &Teardown{Scope: "sources"},
		data:     make(map[string]*Source, 0),
		configDB: cdb,
	}

	return ss, ss.Open()
}

type Sources struct {
	*Teardown

	primary string

	data map[string]*Source

	configDB config.DB
}

func (ss *Sources) Open() error {
	return nil
}

func (ss *Sources) Close() error {
	var lerr error

	for _, ss := range ss.data {
		if err := ss.Close(); err != nil {
			lerr = err
			continue
		}
	}
	if lerr != nil {
		return fmt.Errorf("error/s encountered while closing, last error was: %s", lerr)
	}
	return ss.Teardown.Close()
}

func (ss *Sources) Add(name string, path string) (*Source, error) {
	log.Printf("Adding source %s...", name)

	s, err := NewSource(name, path, ss.configDB)
	ss.data[name] = s
	return s, err
}

func (ss *Sources) AddLazy(name string, completeFn sourceCompleteFunc) (*Source, error) {
	s, err := NewLazySource(name, completeFn, ss.configDB)
	ss.data[name] = s
	return s, err
}

func (ss *Sources) Primary() (bool, *Source, error) {
	if ss.primary == "" {
		return false, &Source{}, errors.New("no primary")
	}
	for n, s := range ss.data {
		if n == ss.primary {
			return true, s, nil
		}
	}
	return false, &Source{}, fmt.Errorf("did not find primary %s", ss.primary)
}

func (ss *Sources) Has(name string) bool {
	_, ok := ss.data[name]
	return ok
}

func (ss *Sources) Get(name string) (bool, *Source, error) {
	if name == "" {
		return ss.Primary()
	}
	return true, ss.data[name], nil
}

func (ss *Sources) MustGet(name string) (s *Source, err error) {
	ok, s, err := ss.Get(name)
	if err != nil {
		return s, err
	}
	if !ok {
		return s, errors.New("no source")
	}
	return s, nil
}

func (ss *Sources) All() []*Source {
	sss := make([]*Source, 0, len(ss.data))

	for _, s := range ss.data {
		sss = append(sss, s)
	}
	return sss
}

func (ss *Sources) Names() []string {
	sss := make([]string, 0, len(ss.data))

	for _, s := range ss.data {
		sss = append(sss, s.Name)
	}
	return sss
}

func (ss *Sources) WhitelistedNames() []string {
	sss := make([]string, 0, len(ss.data))

	for _, s := range ss.data {
		if ss.configDB.IsAcceptedSource(s.Name) {
			sss = append(sss, s.Name)
		}
	}
	return sss
}

func (ss *Sources) ForEach(fn func(*Source) error) error {
	for _, s := range ss.data {
		if err := fn(s); err != nil {
			return err
		}
	}
	return nil
}

func (ss *Sources) SelectPrimary(name string) error {
	log.Printf("Selecting %s as primary source...", name)
	ss.primary = name
	return nil
}
