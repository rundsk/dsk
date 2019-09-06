// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import "net/http"

var (
	Err            = &Error{http.StatusInternalServerError, "Techniker ist informiert"}
	ErrUnsafePath  = &Error{http.StatusBadRequest, "Directory traversal attempt detected!"}
	ErrNotFound    = &Error{http.StatusNotFound, "Not found"}
	ErrNoSuchNode  = &Error{http.StatusNotFound, "No such node"}
	ErrNoSuchAsset = &Error{http.StatusNotFound, "No such asset"}
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
