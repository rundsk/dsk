// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
)

// A document file.
type NodeDoc struct {
	// Absolute path to the document file.
	path string
}

func (d NodeDoc) Hash() ([]byte, error) {
	h := sha1.New()

	f, err := os.Open(d.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = io.Copy(h, f)
	return h.Sum(nil), err
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
//
// The provided set of prefix and node URL will be used to resolve
// relative source URLs and node URLs inside the documents, to
// i.e. make them absolute.
func (d NodeDoc) HTML(treePrefix string, nodeURL string, nodeGet NodeGetter) ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(d.path) {
	case ".md", ".markdown":
		parsed, err := d.parseMarkdown(contents)
		if err != nil {
			return parsed, err
		}
		return d.postprocessHTML(parsed, treePrefix, nodeURL, nodeGet)
	case ".html", ".htm":
		return d.postprocessHTML(contents, treePrefix, nodeURL, nodeGet)
	}
	return nil, fmt.Errorf("Document not in a supported format: %s", prettyPath(d.path))
}

// Raw content of the underlying file.
func (d NodeDoc) Raw() ([]byte, error) {
	return ioutil.ReadFile(d.path)
}

// Parses markdown into HTML and makes relative links absolute, so
// they are more portable.
func (d NodeDoc) parseMarkdown(contents []byte) ([]byte, error) {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags &^ blackfriday.UseXHTML,
	})
	extensions := blackfriday.CommonExtensions |
		blackfriday.Strikethrough | blackfriday.NoEmptyLineBeforeBlock&^
		blackfriday.HeadingIDs&^blackfriday.DefinitionLists

	return blackfriday.Run(
		contents,
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(extensions),
	), nil
}

// Post-processes given HTML after it has been processed by i.e. the
// file-type specific parser.
//
// - Makes the HTML more portable, by turning relative source links
//   into absolute ones.
//
//   Works around buggy AbsolutePrefix feature in blackfriday. We
//   need to have all source URLs absolute as documents can be placed
//   anywhere inside the frontend's URL structure. The workaround can
//   possibly be removed once PR #231 or a change functionally equal
//   to it has been implemented.
//
//   https://github.com/russross/blackfriday/pull/231
//   https://github.com/russross/blackfriday/commit/27ba4cebef7f37e0bb5685e23cb7213cd809f9e8
//   https://github.com/russross/blackfriday/commit/5c12499aa1ddda74561fb899c394f01fd1e8e9e6
//
// - Adds a title atttribute to node links
func (d NodeDoc) postprocessHTML(contents []byte, treePrefix string, nodeURL string, nodeGet NodeGetter) ([]byte, error) {
	var buf bytes.Buffer

	// Append slash to ensure last path element isn't recognized as a file.
	treeBase, err := url.Parse(path.Join(treePrefix, nodeURL) + "/")
	if err != nil {
		return buf.Bytes(), err
	}
	nodeBase, err := url.Parse(nodeURL + "/")
	if err != nil {
		return buf.Bytes(), err
	}

	z := html.NewTokenizer(bytes.NewReader(contents))

	// Helper to get an attribute value from a token.
	attr := func(t html.Token, name string) (bool, int, string) {
		for key, a := range t.Attr {
			if a.Key == name {
				return true, key, a.Val
			}
		}
		return false, 0, ""
	}

	maybeMakeAbsolute := func(t html.Token) (html.Token, error) {
		ok, key, v := attr(t, "src")
		if !ok {
			// No source to change.
			return t, nil
		}
		u, err := url.Parse(v)
		if err != nil {
			return t, err
		}
		if u.IsAbs() {
			return t, nil
		}
		t.Attr[key].Val = treeBase.ResolveReference(u).String()
		return t, nil
	}

	maybeAddTitle := func(t html.Token) (html.Token, error) {
		ok, _, v := attr(t, "title")
		if ok && v != "" {
			// Do not overwrite existing title.
			return t, nil
		}

		ok, _, v = attr(t, "href")
		if !ok {
			// No URL to check at all.
			return t, nil
		}

		u, err := url.Parse(v)
		if err != nil {
			return t, err
		}

		if u.Scheme != "" || u.Host != "" {
			// Doesn't look like a node URL, save the lookup.
			return t, nil
		}
		u = nodeBase.ResolveReference(u)
		v = strings.TrimLeft(u.Path, "/")

		ok, n, err := nodeGet(v)
		if err != nil {
			return t, err
		}
		if !ok {
			return t, nil
		}
		t.Attr = append(t.Attr, html.Attribute{Key: "title", Val: n.Title()})
		t.Attr = append(t.Attr, html.Attribute{Key: "data-node", Val: n.URL()})
		return t, nil
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
			switch t.Data {
			case "img", "video":
				t, err = maybeMakeAbsolute(t)
				if err != nil {
					return buf.Bytes(), err
				}
			case "a":
				t, err = maybeAddTitle(t)
				if err != nil {
					return buf.Bytes(), err
				}
			}
		}
		buf.WriteString(t.String())
	}
}
