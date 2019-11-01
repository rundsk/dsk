// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"testing"
)

func TestRemoveOrderNumbersPrefixed(t *testing.T) {
	if removeOrderNumber("1_cat.jpg") != "cat.jpg" {
		t.Fail()
	}
}

func TestRemoveOrderNumbersLeaveIncorrectlyPrefixed(t *testing.T) {
	if removeOrderNumber("1cat.jpg") != "1cat.jpg" {
		t.Fail()
	}
}

func TestRemoveOrderNumbersLeaveSuffixed(t *testing.T) {
	if removeOrderNumber("cat_1.jpg") != "cat_1.jpg" {
		t.Fail()
	}
	if removeOrderNumber("cat1.jpg") != "cat1.jpg" {
		t.Fail()
	}
}

func TestRemoveOrderNumbersLeavePureNumerical(t *testing.T) {
	if removeOrderNumber("20201224.jpg") != "20201224.jpg" {
		t.Fail()
	}
}
