// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
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
func (d NodeDoc) HTML(treePrefix string, nodeURL string, nodeGet NodeGetter, nodeSource string) ([]byte, error) {
	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	dt, err := NewNodeDocTransformer(treePrefix, nodeURL, nodeGet, nodeSource)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".md", ".markdown":
		components := findComponentsInMarkdown(contents)
		contents = extractComponents(contents, components)

		parsed, err := d.parseMarkdown(contents)
		if err != nil {
			return parsed, err
		}

		parsed = insertComponents(parsed, components)
		return dt.ProcessHTML(parsed)
	case ".html", ".htm":
		return dt.ProcessHTML(contents)
	case ".txt":
		html := fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(contents)))
		return []byte(html), nil
	}
	return nil, fmt.Errorf("unsupported format: %s", d.path)
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
	return nil, fmt.Errorf("unsupported format: %s", d.path)
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

// A headline used in the Table of Contents (ToC)
type TocEntry struct {
	Title    string
	Children []*TocEntry
	Level    int
}

// Generate Table of Contents
func (d NodeDoc) Toc() ([]*TocEntry, error) {
	toc := make([]*TocEntry, 0)

	contents, err := ioutil.ReadFile(d.path)
	if err != nil {
		return nil, err
	}

	document := make([]byte, 0)
	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".md", ".markdown":
		document, _ = d.parseMarkdown(contents)
	case ".html", ".htm":
		document = contents
	}

	r := bytes.NewReader(document)

	doc, _ := html.Parse(r)
	if err != nil {
		return toc, err
	}

	var findHeadlineInDom func(*html.Node)
	findHeadlineInDom = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2" || n.Data == "h3" || n.Data == "h4" || n.Data == "h5") {
			level, _ := strconv.Atoi(strings.TrimPrefix(n.Data, "h"))

			var getTextContent func(*html.Node) string
			getTextContent = func(n *html.Node) string {
				if n.Type == html.TextNode {
					return n.Data
				}
				textOfChildren := ""
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					textOfChildren = textOfChildren + getTextContent(c)
				}
				return textOfChildren
			}
			title := getTextContent(n)

			newEntry := &TocEntry{
				Title:    title,
				Level:    level,
				Children: make([]*TocEntry, 0),
			}

			var insertEntryIntoSubtree func([]*TocEntry) []*TocEntry
			insertEntryIntoSubtree = func(subtree []*TocEntry) []*TocEntry {
				if len(subtree) == 0 {
					subtree = append(subtree, newEntry)
					return subtree
				}
				if subtree[len(subtree)-1].Level >= newEntry.Level {
					subtree = append(subtree, newEntry)
					return subtree
				} else {
					subtree[len(subtree)-1].Children = insertEntryIntoSubtree(subtree[len(subtree)-1].Children)
					return subtree
				}
			}

			toc = insertEntryIntoSubtree(toc)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findHeadlineInDom(c)
			}
		}
	}
	findHeadlineInDom(doc)

	return toc, nil
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

// Extracts components and adds a placeholder instead of it.
func extractComponents(contents []byte, components []*NodeDocComponent) []byte {
	var c string
	var r strings.Builder
	var offset int

	c = string(contents)

	for _, component := range components {
		for i := 0; i < len(c); i++ {
			if i >= component.Position+offset && i < component.Position+component.Length+offset {
				if i == component.Position+offset {
					r.WriteString(component.Placeholder())
				}
				continue
			}
			r.WriteByte(c[i])
		}
		c = r.String()
		r.Reset()
		offset += len(component.Placeholder()) - component.Length
	}
	return []byte(c)
}

// Replaces placeholders with components.
func insertComponents(contents []byte, components []*NodeDocComponent) []byte {
	for _, component := range components {
		contents = bytes.ReplaceAll(contents, []byte(component.Placeholder()), []byte(component.Raw))
	}
	return contents
}
