// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

type DB interface {
	Open() error
	Close() error
	Refresh() error
	GetByEmail(string) (bool, *Author)
}
