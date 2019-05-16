// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "testing"

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
	raw0 := "above\n`<h1>and</h1>`\nbelow"

	result0 := findComponentsInMarkdown([]byte(raw0))
	if len(result0) != 0 {
		t.Errorf("Failed to skip over code, got: %#v", result0)
	}
}

func TestFindInMarkdownExcludeFencedCode(t *testing.T) {
	raw0 := "```\n<h1>hello</h1>\n```"

	result0 := findComponentsInMarkdown([]byte(raw0))
	if len(result0) != 0 {
		t.Errorf("Failed to skip over fenced code, got: %#v", result0)
	}
}
