// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var (
	HTTPErr            = &HTTPError{http.StatusInternalServerError, "Techniker ist informiert"}
	HTTPErrUnsafePath  = &HTTPError{http.StatusBadRequest, "Directory traversal attempt detected!"}
	HTTPErrNoSuchNode  = &HTTPError{http.StatusNotFound, "No such node"}
	HTTPErrNoSuchAsset = &HTTPError{http.StatusNotFound, "No such asset"}
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

type HTTPResponder struct {
	w           http.ResponseWriter
	r           *http.Request
	ContentType string
}

func (re *HTTPResponder) Cached(etag HashGetter) bool {
	hash, err := etag()
	if err != nil {
		log.Print(err)
		return false
	}
	if fmt.Sprintf("%x", hash) == re.r.Header.Get("If-None-Match") {
		re.w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}

func (re *HTTPResponder) Cache(etag HashGetter) {
	hash, err := etag()
	if err != nil {
		log.Print(err)
		return
	}
	re.w.Header().Set("Etag", fmt.Sprintf("%x", hash))
}

func (re *HTTPResponder) OK(data interface{}) {
	re.w.Header().Set("Content-Type", re.ContentType)

	if re.ContentType != "application/json" {
		re.w.WriteHeader(http.StatusOK)
		re.w.Write(data.([]byte))
		return
	}

	jd, jerr := json.Marshal(data)
	if jerr != nil {
		log.Print(jerr)
		re.Error(HTTPErr, jerr)
		return
	}
	re.w.WriteHeader(http.StatusOK)
	re.w.Write(jd)
}

func (re *HTTPResponder) Error(hErr *HTTPError, err error) {
	if hErr.Code != http.StatusNotFound {
		log.Printf("Error (masked) while responding to %s: %s", re.r.URL, err)
	}
	re.w.Header().Set("Content-Type", re.ContentType)
	re.w.WriteHeader(hErr.Code)

	if re.ContentType != "application/json" {
		re.w.Write([]byte(hErr.Message))
		return
	}

	jd, jerr := json.Marshal(hErr)
	if jerr != nil {
		log.Print(jerr)
		re.w.Write([]byte("{}"))
		return
	}
	re.w.Write(jd)
}
