// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const (
	AuthorsConfigBasename = "AUTHORS.txt"
)

func NewAuthorsFromFile(path string) (*Authors, error) {
	as := &Authors{}

	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		return as, err
	}
	return as, as.AddFrom(f)
}

type Authors struct {
	data []*Author
}

type Author struct {
	Email string
	Name  string
}

// Parses given file and adds authors to the internal data.
// Extracts author information from AUTHORS.txt files in mailmap
// format. Currently supports the simple syntax only.
//
// See: https://github.com/git/git/blob/master/Documentation/mailmap.txt
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

func (as Authors) parse(r io.Reader) ([]*Author, error) {
	var parsed []*Author

	lineScanner := bufio.NewScanner(r)

	for lineScanner.Scan() {
		line := lineScanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}
		beginMail := strings.Index(line, "<")

		name := strings.TrimSpace(line[0 : beginMail-1])
		email := strings.TrimSpace(line[beginMail+1 : len(line)-1])
		parsed = append(parsed, &Author{email, name})
	}
	return parsed, nil
}

func (as Authors) Get(email string) *Author {
	for _, a := range as.data {
		if a.Email == email {
			return a
		}
	}
	return nil
}
