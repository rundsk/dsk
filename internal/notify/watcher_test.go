// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package notify

import "testing"

func TestDetectHiddenPathSegmentsInRelativePath(t *testing.T) {
	expected := map[string]bool{
		"xyz/foo/bar":  false,
		".xyz/foo/bar": true,
		"xyz/.foo/bar": true,
		"xyz/foo/.bar": true,
	}
	for path, e := range expected {
		r := anyPathSegmentIsHidden(path)
		if e != r {
			t.Errorf("\nexpected: %v, result: %v, for: %s", e, r, path)
		}
	}
}

func TestDetectHiddenPathSegmentsInAbsolutePath(t *testing.T) {
	expected := map[string]bool{
		"/xyz/foo/bar":  false,
		"/.xyz/foo/bar": true,
		"/xyz/.foo/bar": true,
		"/xyz/foo/.bar": true,
	}
	for path, e := range expected {
		r := anyPathSegmentIsHidden(path)
		if e != r {
			t.Errorf("\nexpected: %v, result: %v, for: %s", e, r, path)
		}
	}
}
