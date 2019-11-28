// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// HashGetter returns a freshly calculated or cached hash.
type HashGetter func() (string, error)

func NewResponder(w http.ResponseWriter, r *http.Request, contentType string, allowOrigin string) *Responder {
	return &Responder{w, r, contentType, allowOrigin}
}

type Responder struct {
	w http.ResponseWriter
	r *http.Request

	ContentType string

	// The value of the Access-Control-Allow-Origin HTTP header to set, if empty
	// the header will remain unset. See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	// for valid values.
	allowOrigin string
}

func (re *Responder) Cached(etag HashGetter) bool {
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

func (re *Responder) Cache(etag HashGetter) {
	hash, err := etag()
	if err != nil {
		log.Print(err)
		return
	}
	re.w.Header().Set("Etag", fmt.Sprintf("%x", hash))
}

func (re *Responder) OK(data interface{}) {
	if re.allowOrigin != "" {
		re.w.Header().Set("Access-Control-Allow-Origin", re.allowOrigin)
		re.w.Header().Set("Vary", "Origin")
	}
	re.w.Header().Set("Content-Type", re.ContentType)

	if re.ContentType != "application/json" {
		re.w.WriteHeader(http.StatusOK)
		re.w.Write(data.([]byte))
		return
	}

	jd, jerr := json.Marshal(data)
	if jerr != nil {
		log.Print(jerr)
		re.Error(Err, jerr)
		return
	}
	re.w.WriteHeader(http.StatusOK)
	re.w.Write(jd)
}

func (re *Responder) Error(hErr *Error, err error) {
	if hErr.Code != http.StatusNotFound {
		log.Printf("Error (masked) while responding to %s: %s", re.r.URL, err)
	}
	if re.allowOrigin != "" {
		re.w.Header().Set("Access-Control-Allow-Origin", re.allowOrigin)
		re.w.Header().Set("Vary", "Origin")
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
