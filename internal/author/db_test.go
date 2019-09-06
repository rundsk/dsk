// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"testing"
)

func TestGetByEmail(t *testing.T) {
	db := &DB{}
	db.Add(&Author{Name: "test", Email: "marius@atelierdisko.de"})

	if ok, _ := db.GetByEmail("marius@atelierdisko.de"); !ok {
		t.Error("failed to lookup by mail")
	}
}
