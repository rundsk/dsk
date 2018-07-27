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
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

// A document file.
type NodeDoc struct {
	// Absolute path to the document file.
	path string
}

// Order is a hint for outside sorting mechanisms.
func (d NodeDoc) Order() uint64 {
	return orderNumber(filepath.Base(d.path))
}

// The document's computed title with any ordering numbers and the
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

	switch strings.ToLower(filepath.Ext(d.path)) {
	case ".md", ".markdown":
		parsed, err := d.parseMarkdown(contents)
		if err != nil {
			return parsed, err
		}
		return d.postprocessHTML(parsed, treePrefix, nodeURL, nodeGet)
	case ".html", ".htm":
		return d.postprocessHTML(contents, treePrefix, nodeURL, nodeGet)
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

// Post-processes given HTML after it has been processed file-type
// specific parsers. See HTML().
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
// - Adds a data-node atttribute to node links
//
// - Adds a data-node attribute to node links containing the node's URL
//
// - Adds dimension attributes to images of nodes.
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
			return t, nil
		}

		u, err := url.Parse(v)
		if err != nil || u.IsAbs() {
			return t, err
		}

		t.Attr[key].Val = treeBase.ResolveReference(u).String()
		return t, nil
	}

	// Works for both for both relative and absolute node URLs.
	maybeAddDataNode := func(t html.Token) (html.Token, error) {
		ok, _, v := attr(t, "href")
		if !ok {
			return t, nil
		}

		u, err := url.Parse(v)
		if err != nil {
			return t, err
		}

		if u.Scheme != "" || u.Host != "" {
			// Doesn't look like a node URL at all, save the lookup.
			return t, nil
		}

		// We look for both "/foo/bar" as well as "foo/bar", whereas
		// the latter will not be considered a relative link when it
		// can be successfully node-looked-up. This is to allow minor
		// human errors, that happen.
		if !strings.HasPrefix(u.Path, "/") {
			u.Path = fmt.Sprintf("/%s", u.Path)
		}
		u = nodeBase.ResolveReference(u)
		v = strings.TrimLeft(u.Path, "/")

		ok, n, err := nodeGet(v)
		if err != nil || !ok {
			return t, err
		}
		t.Attr = append(t.Attr, html.Attribute{Key: "data-node", Val: n.URL()})
		return t, nil
	}

	maybeSize := func(t html.Token) (html.Token, error) {
		ok, _, v := attr(t, "src")
		if !ok {
			// No source to change.
			return t, nil
		}

		u, err := url.Parse(v)
		if err != nil {
			return t, err
		}

		if u.Scheme != "" || u.Host != "" || !strings.HasPrefix(u.Path, "/") {
			// Doesn't look like a node URL at all, save the lookup.
			// The URLs should have already been made absolute in
			// maybeMakeAbsolute().
			return t, nil
		}
		// Let's strip of the routing prefix and split the asset file, so
		// we can look up the asset after node-getting.
		r := strings.TrimPrefix(u.Path, treePrefix)
		nurl := strings.Trim(filepath.Dir(r), "/")

		ok, n, err := nodeGet(nurl)
		if !ok || err != nil {
			return t, err
		}

		a, err := n.Asset(filepath.Base(r))
		if err != nil {
			return t, nil
		}
		ok, w, h, err := a.Dimensions()
		if !ok || err != nil {
			return t, err
		}

		t.Attr = append(t.Attr, html.Attribute{Key: "width", Val: strconv.Itoa(w)})
		t.Attr = append(t.Attr, html.Attribute{Key: "height", Val: strconv.Itoa(h)})
		return t, nil
	}

	z := html.NewTokenizer(bytes.NewReader(contents))
	var isEscaping bool
	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			err := z.Err()

			if err == io.EOF {
				return buf.Bytes(), nil
			}
			return buf.Bytes(), err
		}

		if tt == html.CommentToken || tt == html.DoctypeToken {
			// We don't care about these... ever.
			buf.Write(z.Raw())
			continue
		}

		// By default html.Tokenizer's methods normalize tag names
		// to lower case. As we use custom component tag names in
		// pre-formatted text, we'll need to be sure to keep the
		// casing intact instead.
		//
		// Calling TagName() et al. will modify the underlying
		// slice as returned by Raw(). To prevent this we'll clone
		// the slice.
		raw := append([]byte(nil), z.Raw()...)
		t := z.Token()

		if isEscaping {
			if t.Data == "code" && tt == html.EndTagToken {
				isEscaping = false
			} else {
				// Markdown already escapes HTML entities when they
				// are inside a code block, but doesn't if code was
				// in plain HTML tags. Once we get here we don't know
				// if the code tag was originally generated from
				// Markdown. Ensure we don't double escape in any
				// case.
				buf.WriteString(html.EscapeString(html.UnescapeString(string(raw))))
				continue
			}
		}

		switch {
		case t.Data == "img":
			t, err := maybeMakeAbsolute(t)
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = maybeSize(t)
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(t.String())
		case t.Data == "video" || t.Data == "audio":
			t, err := maybeMakeAbsolute(t)
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(t.String())
		case t.Data == "a" && tt == html.StartTagToken:
			t, err := maybeAddDataNode(t)
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(t.String())
		case t.Data == "code" && tt == html.StartTagToken:
			isEscaping = true
			buf.WriteString(t.String())
		case tt == html.TextToken:
			buf.Write(raw)
		default:
			buf.Write(raw)
		}
	}
}
