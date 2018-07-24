// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

// NodeDoc is a document file.
type NodeDoc struct {
	// Absolute path to the document file.
	path string
}

// Order is a hint for outside sorting mechanisms.
func (d NodeDoc) Order() uint64 {
	return orderNumber(filepath.Base(d.path))
}

// Title of the document and computed with any ordering numbers and the
// extension stripped off, usually for display purposes.
func (d NodeDoc) Title() string {
	base := filepath.Base(d.path)
	return norm.NFC.String(removeOrderNumber(strings.TrimSuffix(base, filepath.Ext(base))))
}

// HTML as parsed from the underlying file. The provided tree prefix
// and node URL will be used to resolve relative source and node URLs
// inside the documents, to i.e. make them absolute.
func (d NodeDoc) HTML(treePrefix string, nodeURL string, nodeGet NodeGetter) ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	dt, err := NewNodeDocTransformer(treePrefix, nodeURL, nodeGet)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".md", ".markdown":
		parsed, err := d.parseMarkdown(contents)
		if err != nil {
			return parsed, err
		}
		return dt.ProcessHTML(parsed)
	case ".html", ".htm":
		return dt.ProcessHTML(contents)
	case ".txt":
		html := fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(contents)))
		return []byte(html), nil
	}
	return nil, fmt.Errorf("Document not in a supported format: %s", prettyPath(d.path))
}

// Raw content of the underlying file.
func (d NodeDoc) Raw() ([]byte, error) {
	return ioutil.ReadFile(d.path)
}

// Parses markdown into HTML.
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
