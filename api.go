// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type API interface {
	// Handlers must be mounted below /api/v{major version}.
	MountHTTPHandlers()
}
