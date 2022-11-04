// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

const (
	maxComponentsPerNodeDoc = 10000
)

type NodeDocComponent struct {
	Name string // i.e. CodeBlock

	Raw      string
	RawInner string

	Level    int // Nesting level
	Position int // Start position inside document.
	Length   int // Length of the component code.
}

func (c *NodeDocComponent) Id() string {
	cleaner := regexp.MustCompile(`\s`)

	content := strings.ToLower(cleaner.ReplaceAllString(c.RawInner, ""))
	content += strconv.Itoa(c.Position)
	return fmt.Sprintf("%x", sha1.Sum([]byte(content)))
}

func (c *NodeDocComponent) Placeholder() string {
	return fmt.Sprintf("dsk+component+%s", c.Id())
}

// TODO: Implement
func findComponentsInHTML(contents []byte) []*NodeDocComponent {
	return make([]*NodeDocComponent, 0)
}

// Will find consider anything that looks like HTML a component. We
// can use this simple approach, as Markdown is the main language in
// HTML can be embedded but will than be ignored.
func findComponentsInMarkdown(contents []byte) []*NodeDocComponent {
	c := string(contents)

	found := make([]*NodeDocComponent, 0)

	var isConsuming bool
	var isLookingForTag bool
	var isCode bool

	var current strings.Builder
	var tagName string
	var openingTag string
	var closingTag string

	var openingTagPosition int

	tagNameRegexp := regexp.MustCompile(`^<([a-zA-Z0-9]+)`)

	for i := 0; i < len(c); i++ {
		if !isConsuming && c[i] == '`' && (i-1 < 0 || c[i-1] != '\\') {
			if i+2 < len(c) && c[i+1] == '`' && c[i+2] == '`' {
				i += 2
			}
			// Set to false and end code, when we've been inside code.
			isCode = !isCode
		}
		if isCode {
			continue
		}
		if isConsuming {
			current.WriteByte(c[i])

			// Decide whether we are ending consumption all together,
			// or just found the initial tag, which we'll later
			// need to check if we need to end consumption.
			if c[i] == '>' {
				if isLookingForTag {
					tagName = tagNameRegexp.FindStringSubmatch(current.String())[1]

					openingTag = current.String()
					closingTag = fmt.Sprintf("</%s>", tagName)

					isLookingForTag = false
					continue
				}

				if strings.Contains(current.String(), closingTag) {
					cmp := &NodeDocComponent{
						Name: tagName,

						Raw:      current.String(),
						RawInner: strings.TrimSuffix(strings.TrimPrefix(current.String(), openingTag), closingTag),

						Level:    0, // Currently finding only top level components.
						Position: openingTagPosition,
						Length:   current.Len(),
					}
					found = append(found, cmp)

					current.Reset()
					tagName = ""
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
		if c[i] == '<' && c[i+1] != '!' {
			current.WriteByte(c[i])

			isConsuming = true
			isLookingForTag = true

			openingTagPosition = i
			continue
		}
	}

	return found
}
