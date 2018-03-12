// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Inspired by https://hackernoon.com/simple-http-middleware-with-go-79a4ad62889b
package main

import (
	"net/http"
)

// Middleware provides a convenient mechanism for filtering HTTP
// requests entering the application. It returns a new handler which
// may perform various operations and should finish by calling the
// next HTTP handler.
type Middleware func(next http.HandlerFunc) http.HandlerFunc

func withNoop(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}
