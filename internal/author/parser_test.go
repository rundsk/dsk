// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	txt := `
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <mariuswilms@mailbox.org>
`
	r := strings.NewReader(txt)

	result, err := parse(r)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("parsed wrong number of authors; expected 2: %v", result)
	}
}

// > Use hash '#' for comments that are either on their own line, or after
//   the email address.
func TestParseComments(t *testing.T) {
	txt := `
# this is a first comment
Christoph Labacher <christoph@atelierdisko.de>
	# this is an indented 2nd comment
Marius Wilms <mariuswilms@mailbox.org> # this is a 3rd comment
# this is the last comment
`
	r := strings.NewReader(txt)

	result, err := parse(r)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("parsed wrong number of authors; expected 2: %v", result)
	}
}
