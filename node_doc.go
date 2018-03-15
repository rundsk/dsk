// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
)

// A markdown document file.
type NodeDoc struct {
	// Absolute path to the file.
	path string
	// The provided prefix will be used to make relative links inside the
	// document absolute.
	URLPrefix string
}

// An order number, as a hint for outside sorting mechanisms.
func (d NodeDoc) Order() uint64 {
	return orderNumber(filepath.Base(d.path))
}

// The document's computed title with any ordering numbers and the
// extension stripped off, usually for display purposes.
func (d NodeDoc) Title() string {
	base := filepath.Base(d.path)
	return removeOrderNumber(strings.TrimSuffix(base, filepath.Ext(base)))
}

// HTML as parsed from the underlying file.
func (d NodeDoc) HTML() ([]byte, error) {
	switch filepath.Ext(d.path) {
	case ".md", ".markdown":
		return d.parseMarkdown()
	}
	return nil, fmt.Errorf("document %s is not in a supported format", d.path)
}

// Raw content of the underlying file.
func (d NodeDoc) Raw() ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// Parses markdown into HTML and makes relative links absolute, so
// they are more portable.
func (d NodeDoc) parseMarkdown() ([]byte, error) {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags &^ blackfriday.UseXHTML,
	})
	extensions := blackfriday.CommonExtensions |
		blackfriday.Strikethrough | blackfriday.NoEmptyLineBeforeBlock&^
		blackfriday.HeadingIDs&^blackfriday.DefinitionLists

	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	parsed := blackfriday.Run(
		contents,
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(extensions),
	)

	// Works around buggy AbsolutePrefix feature in blackfriday. We
	// need to have all URLs absolute as documents can be placed
	// anywhere inside the frontend's URL structure. The workaround
	// can possibly be removed once PR #231 or a change functionally
	// equal to it has been implemented.
	//
	// https://github.com/russross/blackfriday/pull/231
	// https://github.com/russross/blackfriday/commit/27ba4cebef7f37e0bb5685e23cb7213cd809f9e8
	// https://github.com/russross/blackfriday/commit/5c12499aa1ddda74561fb899c394f01fd1e8e9e6
	makeURLsAbsolute := func(parsed []byte, prefix string) ([]byte, error) {
		var buf bytes.Buffer

		// Ensure last path element isn't recognized as a file.
		uBase, err := url.Parse(prefix + "/")
		if err != nil {
			return buf.Bytes(), err
		}

		z := html.NewTokenizer(bytes.NewReader(parsed))

		// Helper to get an attribute value from a token.
		attr := func(t html.Token, name string) (bool, int, string) {
			for key, a := range t.Attr {
				if a.Key == name {
					return true, key, a.Val
				}
			}
			return false, 0, ""
		}

		for {
			tt := z.Next()
			t := z.Token()

			switch tt {
			case html.ErrorToken:
				err := z.Err()

				if err == io.EOF {
					return buf.Bytes(), nil
				}
				return buf.Bytes(), err
			case html.StartTagToken, html.SelfClosingTagToken:
				var ok bool
				var key int
				var v string

				switch t.Data {
				case "img", "video":
					ok, key, v = attr(t, "src")
				}
				if !ok {
					continue
				}
				u, err := url.Parse(v)
				if err != nil {
					return buf.Bytes(), err
				}
				if u.IsAbs() {
					continue
				}
				t.Attr[key].Val = uBase.ResolveReference(u).String()
			}
			buf.WriteString(t.String())
		}
	}

	return makeURLsAbsolute(parsed, d.URLPrefix)
}
