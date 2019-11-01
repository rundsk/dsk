// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetByEmail(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(tmp)

	txt := `
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <mariuswilms@mailbox.org>
`
	ioutil.WriteFile(filepath.Join(tmp, "0"), []byte(txt), 0644)
	db, _ := NewTxtDB(filepath.Join(tmp, "0"))

	if ok, _ := db.GetByEmail("mariuswilms@mailbox.org"); !ok {
		t.Error("failed to lookup by mail")
	}
}
