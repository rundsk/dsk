// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
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

// Parses Markdown into HTML. Adds support for JSX in Markdown similar
// to MDX does but in a very limited way.
func (d NodeDoc) parseMarkdown(contents []byte) ([]byte, error) {
	contents, _ = implyScriptTags(contents)

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

// Similar how Go automatically adds semicolons, implying them, we do
// the same but for JSX and <script>-Tags. Other alternatives though of were:
//
// a) Adding JSX support to the Markdown parser, but this means
//    touching the quite complex parser and maintaining a for that has
//    little chance to be merged back into upstream.
//
// b) Writing our own MDX-Parser, ambitious task, especially as we
//    still consider MDX a moving target. We better wait for a HAST
//    library in Go.
func implyScriptTags(contents []byte) ([]byte, error) {
	c := string(contents)

	var found []string

	var isConsuming bool
	var isLookingForTag bool
	var html strings.Builder
	var openingTag string
	var closingTag string

	for i := 0; i < len(c); i++ {
		if isConsuming {
			html.WriteByte(c[i])

			// Decide whether we are ending consumption all together,
			// or just found the initial tag, which we'll later
			// need to check if we need to end consumption.
			if c[i] == '>' {
				if isLookingForTag {
					re := regexp.MustCompile(`^<[a-zA-Z0-9]+`)

					openingTag = html.String()
					closingTag = fmt.Sprintf("%s>", strings.Replace(re.FindString(openingTag), "<", "</", 1))

					isLookingForTag = false
					continue
				}

				if strings.Contains(html.String(), closingTag) {
					found = append(found, html.String())

					html.Reset()
					openingTag = ""
					closingTag = ""

					isConsuming = false
					continue
				}
			}
			continue
		}

		// Start consumption on anything that remotely doesn't look
		// like Markdown.
		if c[i] == '<' {
			html.WriteByte(c[i])

			isConsuming = true
			isLookingForTag = true
			continue
		}
	}

	for _, f := range found {
		c = strings.Replace(c, f, fmt.Sprintf("<script>%s</script>", f), 1)
	}
	return []byte(c), nil
}
