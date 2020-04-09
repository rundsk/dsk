// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"testing"
)

func TestFindInMarkdown(t *testing.T) {
	raw0 := `
The following visual design has been agreed upon by our team:

<Banner>Hi there!</Banner>

<Warning>Don't do this</Warning>
`
	expected0 := []*NodeDocComponent{
		&NodeDocComponent{Level: 0, Raw: "<Banner>Hi there!</Banner>", Position: 64, Length: 26},
		&NodeDocComponent{Level: 0, Raw: "<Warning>Don't do this</Warning>", Position: 92, Length: 32},
	}
	result0 := findComponentsInMarkdown([]byte(raw0))

	if len(result0) != len(expected0) {
		t.Errorf("Failed number of components mismatch, got: %d", len(result0))
	}
	for k, v := range result0 {
		if v.Position != expected0[k].Position {
			t.Errorf("Failed position, got: %d", v.Position)
		}
	}
}

func TestFindInMarkdownExcludeCode(t *testing.T) {
	raw0 := "`<h1>code</h1>`"
	result0 := findComponentsInMarkdown([]byte(raw0))

	if len(result0) != 0 {
		t.Errorf("Failed to skip over code, got: %#v", result0)
	}

	raw1 := "above\n`<h1>and</h1>`\nbelow"
	result1 := findComponentsInMarkdown([]byte(raw1))

	if len(result1) != 0 {
		t.Errorf("Failed to skip over code, got: %#v", result1)
	}

	raw2 := "Usually you'll be using fenced code blocks when authoring Markdown.\nThese get automatically converted to a `<CodeBlock>`."
	result2 := findComponentsInMarkdown([]byte(raw2))

	if len(result2) != 0 {
		t.Errorf("Failed to skip over code, got: %#v", result2)
	}

	raw3 := "...to a `<CodeBlock>`\n<CodeBlock>hello world!</CodeBlock>"
	result3 := findComponentsInMarkdown([]byte(raw3))
	expected3 := []*NodeDocComponent{
		&NodeDocComponent{Level: 0, Raw: "<CodeBlock>hello world!</CodeBlock>", Position: 22, Length: 35},
	}
	if len(result3) != len(expected3) {
		t.Errorf("Failed number of components mismatch, got: %d", len(result3))
	}
	for k, v := range result3 {
		if v.Position != expected3[k].Position {
			t.Errorf("Failed position, got: %d", v.Position)
		}
	}
}

func TestFindInMarkdownExcludeComment(t *testing.T) {
	raw0 := `
# Test

<!-- comment -->

<Component>1</Component>

<!-- comment -->

<Component>2</Component>
`
	result0 := findComponentsInMarkdown([]byte(raw0))

	if len(result0) != 2 {
		t.Errorf("Failed to skip over comment, got: %#v", result0)
	}

	if result0[0].Raw != "<Component>1</Component>" || result0[1].Raw != "<Component>2</Component>" {
		t.Errorf("Failed to skip over comment, got: %#v", result0)
	}
}

func TestFindInMarkdownExcludeFencedCode(t *testing.T) {
	raw0 := "```\n<h1>hello</h1>\n```"
	result0 := findComponentsInMarkdown([]byte(raw0))

	if len(result0) != 0 {
		t.Errorf("Failed to skip over fenced code, got: %#v", result0)
	}
}

func TestFindInMarkdownLiteralFencedCodeFragment(t *testing.T) {
	raw0 := "use \\`\\`\\` ...to a `<CodeBlock>`\n<CodeBlock>hello world!</CodeBlock>"
	result0 := findComponentsInMarkdown([]byte(raw0))
	expected0 := []*NodeDocComponent{
		&NodeDocComponent{Level: 0, Raw: "<CodeBlock>hello world!</CodeBlock>", Position: 33, Length: 35},
	}
	if len(result0) != len(expected0) {
		t.Errorf("Failed number of components mismatch, got: %d", len(result0))
	}
	for k, v := range result0 {
		if v.Position != expected0[k].Position {
			t.Errorf("Failed position, got: %d", v.Position)
		}
	}
}
