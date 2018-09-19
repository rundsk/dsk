// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// AuthorsConfigBasename is the canonical name of the file, we
	// expect the authors text database to live in.
	AuthorsConfigBasename = "AUTHORS.txt"
)

func NewAuthors(path string) *Authors {
	return &Authors{
		path: filepath.Join(path, AuthorsConfigBasename),
		data: make([]*Author, 0),
	}
}

// Authors allows to Extract author information from AUTHORS.txt files
// in mailmap format. Currently supports the simple syntax only.
//
// See: https://github.com/git/git/blob/master/Documentation/mailmap.txt
type Authors struct {
	path string
	data []*Author
}

type Author struct {
	Email string
	Name  string
}

func (as *Authors) Sync() error {
	as.data = make([]*Author, 0)

	if _, err := os.Stat(as.path); os.IsNotExist(err) {
		return nil
	}
	f, err := os.Open(as.path)
	defer f.Close()

	if err != nil {
		return err
	}
	return as.AddFrom(f)
}

// Add single author item to the internal data slice.
func (as *Authors) Add(a *Author) {
	as.data = append(as.data, a)
}

// Parses given file and adds authors to the internal data.
func (as *Authors) AddFrom(r io.Reader) error {
	parsed, err := as.parse(r)
	if err != nil {
		return err
	}
	for _, a := range parsed {
		as.data = append(as.data, a)
	}
	return nil
}

// Parse lines looking like this:
//   Proper Name <commit@email.xx>
//   # this is a comment
//   Proper Name <commit@email.xx> # inline comment
func (as Authors) parse(r io.Reader) ([]*Author, error) {
	var parsed []*Author

	lineScanner := bufio.NewScanner(r)

	for lineScanner.Scan() {
		line := strings.TrimSpace(lineScanner.Text())

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		inlineComment := strings.Index(line, "#")

		beginMail := strings.Index(line, "<")
		if beginMail == -1 || (inlineComment != -1 && inlineComment < beginMail) {
			return parsed, fmt.Errorf("expected opening angle bracket in line '%s'", line)
		}
		endMail := strings.LastIndex(line, ">")
		if endMail == -1 || (inlineComment != -1 && inlineComment < endMail) {
			return parsed, fmt.Errorf("expected closing angle bracket in line '%s'", line)
		}

		name := strings.TrimSpace(line[0 : beginMail-1])
		email := strings.TrimSpace(line[beginMail+1 : len(line)-1])
		parsed = append(parsed, &Author{email, name})
	}
	return parsed, nil
}

func (as Authors) Get(email string) (ok bool, a *Author, err error) {
	for _, a := range as.data {
		if a.Email == email {
			return true, a, nil
		}
	}
	return false, &Author{}, nil
}
