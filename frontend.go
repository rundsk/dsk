// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dev

package main

import (
	"net/http"
	"os"
)

var assets http.FileSystem

func init() {
	env, ok := os.LookupEnv("FRONTEND")
	if !ok {
		assets = http.Dir("./frontend/build")
	} else {
		assets = http.Dir(env)
	}
}
