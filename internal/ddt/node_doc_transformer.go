// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"bytes"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var (
	// HTML attributes that may contain URLs. This should not include "src".
	genericURLAttrNames = []string{"annotate"}
)

// NewNodeDocTransformer returns an initialized NodeDocTransformer, it'll
// derive values for nodeBase and treeBase from the treePrefix and nodeURL.
func NewNodeDocTransformer(treePrefix string, nodeURL string, nodeGet NodeGetter, nodeSource string) (*NodeDocTransformer, error) {
	dt := &NodeDocTransformer{
		treePrefix: treePrefix,
		nodeSource: nodeSource,
		nodeURL:    nodeURL,
		nodeGet:    nodeGet,
	}
	// Append slash to ensure last path element isn't recognized as a file.
	treeBase, err := url.Parse(path.Join(treePrefix, nodeURL) + "/")
	if err != nil {
		return dt, err
	}
	dt.treeBase = treeBase

	nodeBase, err := url.Parse(nodeURL + "/")
	if err != nil {
		return dt, err
	}
	dt.nodeBase = nodeBase

	return dt, err
}

// NodeDocTransformer post-processes given HTML after it has been processed file-type
// specific parsers. See NodeDoc.HTML().
//
// Makes the HTML more portable, by turning relative source links
// into absolute ones. We need to have all source URLs absolute
// as documents can be placed anywhere inside the frontend's URL
// structure.
//
// All references to nodes and node assets are discovered and
// re-constructed as an absolute URL, using the canonical node URL.
//
// All other relative links are made absolute using treeBase.
//
// For elements referencing a node, "data-node" attribute containing
// the node's ref-URL is added. For elements referencing a node asset,
// a "data-node" attribute containing the node's ref-URL is added and
// a "data-node-asset" attribute with the name of the asset is added.
//
// Dimension attributes are added to images of nodes.
//
// HTML inside <code> tags is escaped while preventing double escaping.
type NodeDocTransformer struct {
	// i.e. /tree
	treePrefix string

	// i.e. /tree/foo/bar/
	treeBase *url.URL

	// i.e. foo/bar
	nodeURL string

	// i.e. foo/bar/
	nodeBase *url.URL

	// Allows us to lookup nodes by their ref-URL.
	nodeGet NodeGetter

	// nodeSource is the name of a plex.Source
	nodeSource string
}

// ProcessHTML is the main entry point.
func (dt NodeDocTransformer) ProcessHTML(contents []byte) ([]byte, error) {
	var buf bytes.Buffer

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

		srcok, _, _ := dt.attr(t, "src")
		genok, genkeys := dt.attrs(t, genericURLAttrNames)

		// Order matters: maybeAddDateNode should come first.
		switch {
		case t.Data == "img":
			t, err := dt.maybeAddDataNode(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = dt.maybeMakeAbsolute(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = dt.maybeSize(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(html.UnescapeString(t.String()))
		case t.Data == "video" || t.Data == "audio":
			t, err := dt.maybeAddDataNode(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = dt.maybeMakeAbsolute(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(html.UnescapeString(t.String()))
		case t.Data == "a" && tt == html.StartTagToken:
			t, err := dt.maybeAddDataNode(t, "href")
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = dt.maybeMakeAbsolute(t, "href")
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(html.UnescapeString(t.String()))
		case srcok:
			t, err := dt.maybeAddDataNode(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			t, err = dt.maybeMakeAbsolute(t, "src")
			if err != nil {
				return buf.Bytes(), err
			}
			buf.WriteString(html.UnescapeString(t.String()))
		case genok:
			for _, genkey := range genkeys {
				t, err := dt.maybeMakeAbsolute(t, t.Attr[genkey].Key)
				if err != nil {
					return buf.Bytes(), err
				}
				buf.WriteString(html.UnescapeString(t.String()))
			}
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

// Adds a data-node attribute with the ref-URL of the node to
// the referencing element, when the URL can be resolved and
// successfully looked up as a ddt.
//
// Given "bar" is a node under "foo" and "/tree" is the tree
// prefix, all of the following will be resolved to the "foo/bar"
// node whereever they are found.
//
//   /foo/bar
//   foo/bar
//   /tree/foo/bar
//
// While viewing the "baz" node, which is another node under "foo":
//   ../bar
//
// While viewing the "foo" node:
//   ./bar
//   bar
//
// While viewing the "bar" node, the following will resolve to itself:
// 	 ./
//	 .
//   <empty string>
//
// The references can either target nodes or node assets, both are
// supported:
//  /foo/bar/cat.jpg
func (dt NodeDocTransformer) maybeAddDataNode(t html.Token, attrName string) (html.Token, error) {
	ok, _, v := dt.attr(t, attrName)
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

	// Blindly try to lookup and see it already succeeds, this enables
	// support for both "/foo/bar" and "foo/bar", that is when the leading
	// slash has been forgotten.
	var okdn bool
	var okdna bool
	var dn string
	var dna string

	okdn, dn, okdna, dna = dt.discoverNodeInfo(u)
	if !okdn {
		// Retry while making the URL absolute.
		okdn, dn, okdna, dna = dt.discoverNodeInfo(dt.nodeBase.ResolveReference(u))
		if !okdn {
			return t, nil
		}
	}

	if okdn {
		t.Attr = append(t.Attr, html.Attribute{Key: "data-node", Val: dn})
	}
	if okdna {
		t.Attr = append(t.Attr, html.Attribute{Key: "data-node-asset", Val: dna})
	}
	return t, nil
}

// If it discovers a "data-node" attribute and additionally a
// "data-node-asset" attribute it will always use its information to
// make the link absolute, even if it's already absolute.
//
// When query string or fragment are present in the original URL they
// will be added to the resulting URL generated from the node URL.
//
// Other relative links are made absolute using the treeBase.
func (dt NodeDocTransformer) maybeMakeAbsolute(t html.Token, attrName string) (html.Token, error) {
	ok, key, v := dt.attr(t, attrName)
	if !ok {
		return t, nil
	}

	u, err := url.Parse(v)
	if err != nil {
		return t, err
	}

	// References to nodes or its assets are always made absolute.
	ok, _, dn := dt.attr(t, "data-node")
	if ok {
		dnu := url.URL{
			// Path is augmented with the asset when one is found.
			Path: path.Join(dt.treePrefix, dn),
			// Transfer fragment from original URL.
			Fragment: u.Fragment,
		}
		// We start with the query values of the original URL, as we
		// want to copy them over to the resulting URL.
		q := u.Query()
		// We programatically add to the query, and add
		// it store it on the dnu after we are finished.
		q.Set("v", dt.nodeSource)
		dnu.RawQuery = q.Encode()

		ok, _, dna := dt.attr(t, "data-node-asset")
		if ok {
			dnu.Path = path.Join(dnu.Path, dna)
		}

		t.Attr[key].Val = dnu.String()
		return t, nil
	}

	// Don't touch absolute possibly external links.
	if u.IsAbs() {
		return t, err
	}
	t.Attr[key].Val = dt.treeBase.ResolveReference(u).String()
	return t, nil
}

// Works only for node assets that are images.
func (dt NodeDocTransformer) maybeSize(t html.Token, attrName string) (html.Token, error) {
	ok, _, dn := dt.attr(t, "data-node")
	if !ok {
		return t, nil
	}
	ok, n, err := dt.nodeGet(dn)
	if err != nil {
		return t, err
	}

	ok, _, dna := dt.attr(t, "data-node-asset")
	if !ok {
		return t, nil
	}
	ok, a, err := n.Asset(dna)
	if !ok || err != nil {
		return t, err
	}
	ok, w, h, err := a.Dimensions()
	if !ok || err != nil {
		return t, err
	}

	t.Attr = append(t.Attr, html.Attribute{Key: "width", Val: strconv.Itoa(w)})
	t.Attr = append(t.Attr, html.Attribute{Key: "height", Val: strconv.Itoa(h)})
	return t, nil
}

// Helper to get an attribute value from a token.
func (dt NodeDocTransformer) attr(t html.Token, name string) (bool, int, string) {
	for key, a := range t.Attr {
		if a.Key == name {
			return true, key, a.Val
		}
	}
	return false, 0, ""
}

// Returns attribute keys for matched names.
func (dt NodeDocTransformer) attrs(t html.Token, names []string) (ok bool, keys []int) {
	for key, a := range t.Attr {
		for _, name := range names {
			if a.Key == name {
				keys = append(keys, key)
			}
		}
	}
	return len(keys) != 0, keys
}

// Tries to lookup path of the URL as a node, if that fails tries to
// lookup as node an node asset. Returns string values usuable for
// data attributes.
//
// Using ok return values, as the URL for the root node is an empty
// string, and thus isn't usable to check if the discovery succeeded.
func (dt NodeDocTransformer) discoverNodeInfo(u *url.URL) (bool, string, bool, string) {
	tu := strings.Trim(u.Path, "/")
	if (tu == "" || tu == ".") && u.Path != "." {
		tu = "./"
	}
	if tu == ".." {
		tu = "../"
	}
	ok, n, _ := dt.nodeGet(tu)
	if ok {
		return true, n.URL(), false, ""
	}
	// Retry while removing the last part of the URL, it might
	// be an asset of the ddt.
	ok, n, _ = dt.nodeGet(strings.TrimLeft(path.Dir(tu), "/"))
	if !ok {
		return false, "", false, ""
	}
	// We've found a node but cannot be sure that the asset really
	// is part of the node, let's check that.
	ok, a, _ := n.Asset(path.Base(u.Path))
	if ok {
		return true, n.URL(), true, a.Name()
	}
	// We'll ignore invalid assets on valid nodes for now and keep on going.
	return false, n.URL(), false, ""
}
