// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func TestParseAuthors(t *testing.T) {
	txt := `
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <marius@atelierdisko.de>
`
	r := strings.NewReader(txt)

	as := &Authors{}
	result, _ := as.parse(r)

	if len(result) != 2 {
		t.Errorf("parsed wrong number of authors; expected 2: %v", result)
	}
}

func TestLookupAuthor(t *testing.T) {
	txt := `
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <marius@atelierdisko.de>
`
	r := strings.NewReader(txt)

	as := &Authors{}
	as.AddFrom(r)

	if as.Get("marius@atelierdisko.de") == nil {
		t.Error("failed to lookup by mail")
	}
}
