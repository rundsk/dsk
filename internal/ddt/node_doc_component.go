// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

const (
	maxComponentsPerNodeDoc = 10000
)

type NodeDocComponent struct {
	Id       int // Unique ID
	Raw      string
	Level    int // Nesting level
	Position int // Start position inside document.
	Length   int // Length of the component code.
}

func NewNodeDocComponent(raw string, level int, position int) *NodeDocComponent {
	return &NodeDocComponent{
		Id:       rand.Intn(maxComponentsPerNodeDoc),
		Raw:      raw,
		Level:    level,
		Position: position,
		Length:   len(raw),
	}
}

func (c *NodeDocComponent) Placeholder() string {
	return fmt.Sprintf("dsk+component+%d", c.Id)
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
	var openingTag string
	var closingTag string

	var openingTagPosition int

	for i := 0; i < len(c); i++ {
		if c[i] == '`' && (i-1 < 0 || c[i-1] != '\\') {
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
					re := regexp.MustCompile(`^<[a-zA-Z0-9]+`)

					openingTag = current.String()
					closingTag = fmt.Sprintf("%s>", strings.Replace(re.FindString(openingTag), "<", "</", 1))

					isLookingForTag = false
					continue
				}

				if strings.Contains(current.String(), closingTag) {
					found = append(found, NewNodeDocComponent(
						current.String(),
						0, // Currently finding only top level components.
						openingTagPosition,
					))

					current.Reset()
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
