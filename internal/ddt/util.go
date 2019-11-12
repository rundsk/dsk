// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"strconv"
)

// Finds an order number embedded into given path segment and
// returns it. If none is found, returns 0.
func orderNumber(segment string) uint64 {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) > 2 {
		parsed, _ := strconv.ParseUint(s[0], 10, 64)
		return parsed
	}
	return 0
}

// Removes order numbers from path segment, if present.
func removeOrderNumber(segment string) string {
	s := NodePathTitleRegexp.FindStringSubmatch(segment)

	if len(s) == 0 {
		return segment
	}
	if len(s) > 2 {
		return s[2]
	}
	return s[1]
}
