// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package meta

import (
	"time"
)

// DB stores information about files and directories.
type DB interface {
	Modified(string) (time.Time, error)
	Refresh() error
}
