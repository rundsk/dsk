// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"testing"
)

func TestAssetTitlesWithDecomposedFilename(t *testing.T) {
	a := NewNodeAsset("/bar/Cafe\u0301.json", "bar/cafe.json", nil)
	if a.Title() != "Café" {
		t.Errorf("failed to decode name, got %v", a.Title())
	}
}
