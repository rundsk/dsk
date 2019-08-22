// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

const (
	protectOpeningScriptTag = "<script>'|dsk|'"
	protectClosingScriptTag = "'|dsk|'</script>"
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

// Name is the basename of the file without its order number.
func (d NodeDoc) Name() string {
	return removeOrderNumber(norm.NFC.String(filepath.Base(d.path)))
}

// Title of the document and computed with any ordering numbers and the
// extension stripped off, usually for display purposes.
// We normalize the title string to make sure all special characters
// are represented in their composed form. For more on this topic see the
// docblock of Node.Title().
func (d NodeDoc) Title() string {
	base := norm.NFC.String(filepath.Base(d.path))
	return removeOrderNumber(strings.TrimSuffix(base, filepath.Ext(base)))
}

// HTML as parsed from the underlying file. The provided tree prefix
// and node URL will be used to resolve relative source and node URLs
// inside the documents, to i.e. make them absolute.
func (d NodeDoc) HTML(treePrefix string, nodeURL string, nodeGet NodeGetter) ([]byte, error) {
	start := time.Now()
	defer log.Printf("Rendered document %s in %s", prettyPath(d.path), time.Since(start))

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
		contents = addComponentProtection(contents, findComponentsInMarkdown(contents))

		parsed, err := d.parseMarkdown(contents)
		if err != nil {
			return parsed, err
		}

		parsed = removeComponentProtection(parsed)
		return dt.ProcessHTML(parsed)
	case ".html", ".htm":
		return dt.ProcessHTML(contents)
	case ".txt":
		html := fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(contents)))
		return []byte(html), nil
	}
	return nil, fmt.Errorf("Document not in a supported format: %s", prettyPath(d.path))
}

// Text converted from original file format.
func (d NodeDoc) CleanText() ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	policy := bluemonday.StrictPolicy()

	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".txt":
		return contents, nil
	case ".md", ".markdown":
		c, _ := d.parseMarkdown(contents)
		return policy.SanitizeBytes(c), nil
	case ".html", ".htm":
		return policy.SanitizeBytes(contents), nil
	}
	return nil, fmt.Errorf("Document not in a supported format: %s", prettyPath(d.path))
}

// Raw content of the underlying file.
func (d NodeDoc) Raw() ([]byte, error) {
	return ioutil.ReadFile(d.path)
}

// Components as found in the raw document.
func (d NodeDoc) Components() ([]*NodeDocComponent, error) {
	components := make([]*NodeDocComponent, 0)

	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return components, err
	}

	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".md", ".markdown":
		return findComponentsInMarkdown(contents), nil
	case ".html", ".htm":
		return findComponentsInHTML(contents), nil
	}
	return components, nil
}

// Parses Markdown into HTML.
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

// Used to protect component code from the Markdown parser. The
// implied script tags can be removed once the Markdown has been
// transformed into HTML. Similar to how Go automatically implies
// semicolons.
func addComponentProtection(contents []byte, components []*NodeDocComponent) []byte {
	var c string
	var r strings.Builder
	var offset int

	c = string(contents)

	for _, component := range components {
		for i := 0; i < len(c); i++ {
			if i == component.Position+offset {
				r.WriteString(protectOpeningScriptTag)

			} else if i == component.Position+component.Length+offset {
				r.WriteString(protectClosingScriptTag)
			}
			r.WriteByte(c[i])
		}
		if len(c) == component.Position+component.Length+offset {
			r.WriteString(protectClosingScriptTag)
		}
		c = r.String()
		r.Reset()
		offset += len(protectOpeningScriptTag) + len(protectClosingScriptTag)
	}
	return []byte(c)
}

func removeComponentProtection(contents []byte) []byte {
	contents = bytes.ReplaceAll(contents, []byte(protectOpeningScriptTag), []byte{})
	contents = bytes.ReplaceAll(contents, []byte(protectClosingScriptTag), []byte{})

	return contents
}
