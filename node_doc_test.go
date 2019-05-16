// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDocTitlesWithDecomposedFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "01-Cafe\u0301.md")
	ioutil.WriteFile(doc0, []byte(""), 0666)

	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}

	docs, err := node.Docs()
	if err != nil {
		t.Errorf("can’t read docs")
	}

	if docs[0].Title() != "Café" {
		t.Errorf("failed to decode file name, got %v", docs[0].Title())
	}
}

func TestAddComponentProtection(t *testing.T) {
	raw0 := `
<CodeBlock title="Example">
  echo "GREETINGS PROFESSOR FALKEN."
</CodeBlock>
`
	expected0 := `
<script>'|dsk|'<CodeBlock title="Example">
  echo "GREETINGS PROFESSOR FALKEN."
</CodeBlock>'|dsk|'</script>
`
	components0 := []*NodeDocComponent{
		&NodeDocComponent{
			Raw:      raw0[1 : len(raw0)-1],
			Length:   len(raw0) - 2,
			Position: 1,
		},
	}

	result0 := addComponentProtection([]byte(raw0), components0)
	if string(result0) != expected0 {
		t.Errorf("Failed, got: %s", result0)
	}

	raw1 := `
Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
`
	expected1 := `
Yellow and <script>'|dsk|'<ColorSwatch>green</ColorSwatch>'|dsk|'</script> are the colors of spring.
`
	components1 := []*NodeDocComponent{
		&NodeDocComponent{
			Raw:      "<ColorSwatch>green</ColorSwatch>",
			Length:   len("<ColorSwatch>green</ColorSwatch>"),
			Position: 12,
		},
	}

	result1 := addComponentProtection([]byte(raw1), components1)
	if string(result1) != expected1 {
		t.Errorf("Failed, got: %s", result1)
	}

	raw2 := `
The following visual design has been agreed upon by our team:

<Banner>Hi there!</Banner>

<Warning>Don't do this</Warning>
`
	expected2 := `
The following visual design has been agreed upon by our team:

<script>'|dsk|'<Banner>Hi there!</Banner>'|dsk|'</script>

<script>'|dsk|'<Warning>Don't do this</Warning>'|dsk|'</script>
`
	components2 := []*NodeDocComponent{
		&NodeDocComponent{Level: 0, Raw: "<Banner>Hi there!</Banner>", Position: 64, Length: 26},
		&NodeDocComponent{Level: 0, Raw: "<Warning>Don't do this</Warning>", Position: 92, Length: 32},
	}
	result2 := addComponentProtection([]byte(raw2), components2)

	if string(result2) != expected2 {
		t.Errorf("Failed, got: %s", result2)
	}
}

func TestRemoveComponentProtection(t *testing.T) {
	raw0 := `
Yellow and <script>'|dsk|'<ColorSwatch>green</ColorSwatch>'|dsk|'</script> are the colors of spring.
`
	expected0 := `
Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
`

	result0 := removeComponentProtection([]byte(raw0))
	if string(result0) != expected0 {
		t.Errorf("Failed, got: %s", result0)
	}
}

func TestComponentIsLeftUntouchedInHTMLDocument(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	get := func(url string) (bool, *Node, error) {
		return false, &Node{}, nil
	}

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "readme.html")
	raw0 := `
<dsk-code-block>
  echo "GREETINGS PROFESSOR FALKEN."	
</dsk-code-block>
`
	expected0 := `
<dsk-code-block>
  echo "GREETINGS PROFESSOR FALKEN."	
</dsk-code-block>
`
	ioutil.WriteFile(doc0, []byte(raw0), 0666)

	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
	docs, _ := node.Docs()

	html0, _ := docs[0].HTML("/tree", "foo/bar", get)

	if string(html0) != expected0 {
		t.Errorf("Component markup does not look like expected, got: %s", html0)
	}
}

func TestComponentsAreLeftUntouchedInMarkdownDocument(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	get := func(url string) (bool, *Node, error) {
		return false, &Node{}, nil
	}

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "readme.md")
	raw0 := `
# Visual Design

The following visual design has been agreed upon by our team:

<Banner>Hi there!</Banner>

<Warning>Don't do this</Warning>

hello

<CodeBlock title="test">
	<h1>Hello Headline</h1>
</CodeBlock>
`
	expected0 := `<h1>Visual Design</h1>

<p>The following visual design has been agreed upon by our team:</p>

<Banner>Hi there!</Banner>

<Warning>Don't do this</Warning>

<p>hello</p>

<CodeBlock title="test">
	<h1>Hello Headline</h1>
</CodeBlock>
`
	ioutil.WriteFile(doc0, []byte(raw0), 0666)

	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
	docs, _ := node.Docs()

	html0, _ := docs[0].HTML("/tree", "foo/bar", get)

	if string(html0) != expected0 {
		t.Errorf("Component markup does not look like expected, got: %s", html0)
	}
}
