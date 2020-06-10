// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocTitlesWithDecomposedFilenames(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "01-Cafe\u0301.md")
	ioutil.WriteFile(doc0, []byte(""), 0666)

	node := &Node{root: tmp, Path: filepath.Join(tmp, "foo")}

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
dsk+component+23
`
	components0 := []*NodeDocComponent{
		&NodeDocComponent{
			Id:       23,
			Raw:      raw0[1 : len(raw0)-1],
			Length:   77,
			Position: 1,
		},
	}

	result0 := extractComponents([]byte(raw0), components0)
	if string(result0) != expected0 {
		t.Errorf("Failed, got: %s", result0)
	}

	raw1 := `
Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
`
	expected1 := `
Yellow and dsk+component+23 are the colors of spring.
`
	components1 := []*NodeDocComponent{
		&NodeDocComponent{
			Id:       23,
			Raw:      "<ColorSwatch>green</ColorSwatch>",
			Length:   len("<ColorSwatch>green</ColorSwatch>"),
			Position: 12,
		},
	}

	result1 := extractComponents([]byte(raw1), components1)
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

dsk+component+0

dsk+component+1
`
	components2 := []*NodeDocComponent{
		&NodeDocComponent{Id: 0, Level: 0, Raw: "<Banner>Hi there!</Banner>", Position: 64, Length: 26},
		&NodeDocComponent{Id: 1, Level: 0, Raw: "<Warning>Don't do this</Warning>", Position: 92, Length: 32},
	}
	result2 := extractComponents([]byte(raw2), components2)

	if string(result2) != expected2 {
		t.Errorf("Failed, got: %s", result2)
	}
}

func TestAddComponentProtectionOnLastLine(t *testing.T) {
	raw0 := `
<Banner title="Banner" type="warning">Use banners to highlight things people shouldn’t miss.</Banner>

<Banner title="Banner" type="warning">Use banners to highlight things people shouldn’t miss.</Banner>`
	expected0 := `
dsk+component+0

dsk+component+1`

	components0 := []*NodeDocComponent{
		&NodeDocComponent{Id: 0, Level: 0, Raw: "<Banner title=\"Banner\" type=\"warning\">Use banners to highlight things people shouldn’t miss.</Banner>", Position: 1, Length: 103},
		&NodeDocComponent{Id: 1, Level: 0, Raw: "<Banner title=\"Banner\" type=\"warning\">Use banners to highlight things people shouldn’t miss.</Banner>", Position: 106, Length: 103},
	}

	result0 := extractComponents([]byte(raw0), components0)
	if string(result0) != expected0 {
		t.Errorf("Failed, got: %s", result0)
	}
}

func TestRemoveComponentProtection(t *testing.T) {
	raw0 := `
Yellow and dsk+component+23 are the colors of spring.
`
	expected0 := `
Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
`
	components0 := []*NodeDocComponent{
		&NodeDocComponent{
			Id:       23,
			Raw:      "<ColorSwatch>green</ColorSwatch>",
			Length:   len("<ColorSwatch>green</ColorSwatch>") - 2,
			Position: 1,
		},
	}

	result0 := insertComponents([]byte(raw0), components0)
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

	node := &Node{root: tmp, Path: filepath.Join(tmp, "foo")}
	docs, _ := node.Docs()

	html0, _ := docs[0].HTML("/tree", "foo/bar", get, "test")

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

<p><Banner>Hi there!</Banner></p>

<p><Warning>Don't do this</Warning></p>

<p>hello</p>

<p><CodeBlock title="test">
	<h1>Hello Headline</h1>
</CodeBlock></p>
`
	ioutil.WriteFile(doc0, []byte(raw0), 0666)

	node := &Node{root: tmp, Path: filepath.Join(tmp, "foo")}
	docs, _ := node.Docs()

	html0, _ := docs[0].HTML("/tree", "foo/bar", get, "test")

	if string(html0) != expected0 {
		t.Errorf("Component markup does not look like expected, got: %s", html0)
	}
}

func TestAddRemoveComponentProtectionSymmetry(t *testing.T) {
	raw0 := `
<Banner title="Banner" type="warning">Use banners to highlight things people shouldn’t miss.</Banner>

<Banner title="Banner" type="warning">Use banners to highlight things people shouldn’t miss.</Banner>
`

	components0 := findComponentsInMarkdown([]byte(raw0))
	components0[0].Id = 0
	components0[1].Id = 1
	expectedComponents0 := []*NodeDocComponent{
		&NodeDocComponent{Id: 0, Level: 0, Raw: "<Banner title=\"Banner\" type=\"warning\">Use banners to highlight things people shouldn’t miss.</Banner>", Position: 1, Length: 103},
		&NodeDocComponent{Id: 1, Level: 0, Raw: "<Banner title=\"Banner\" type=\"warning\">Use banners to highlight things people shouldn’t miss.</Banner>", Position: 106, Length: 103},
	}
	if len(components0) != len(expectedComponents0) {
		t.Errorf("Failed number of components mismatch, got: %d", len(components0))
	}
	for k, v := range components0 {
		if v.Position != expectedComponents0[k].Position {
			t.Errorf("Failed position, got: %d", v.Position)
		}
		if v.Length != expectedComponents0[k].Length {
			t.Errorf("Failed length, got: %d", v.Length)
		}
	}

	added0 := extractComponents([]byte(raw0), components0)
	addedExpected0 := `
dsk+component+0

dsk+component+1
`
	if string(added0) != addedExpected0 {
		t.Errorf("Failed, got: %s", added0)
	}

	removed0 := insertComponents(added0, components0)
	removedExpected0 := raw0
	if string(removed0) != removedExpected0 {
		t.Errorf("Failed, got: %s", removed0)
	}
}

func TestGenerateToCForMarkdownDocument(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "tree")
	defer os.RemoveAll(tmp)

	node0 := filepath.Join(tmp, "foo")
	os.Mkdir(node0, 0777)

	doc0 := filepath.Join(node0, "readme.md")
	raw0 := `
# Heading 1
## Heading 2
### Heading 3
## Heading 2
# Heading 1
#### Heading 4
## Heading 2
#### Heading 4
# Heading 1
`

	ioutil.WriteFile(doc0, []byte(raw0), 0666)

	node := &Node{root: tmp, Path: filepath.Join(tmp, "foo")}
	docs, _ := node.Docs()

	toc0, _ := docs[0].Toc()

	expected0 := []*TocEntry{&TocEntry{
		Title: "Heading 1",
		Level: 1,
		Children: []*TocEntry{&TocEntry{
			Title: "Heading 2",
			Level: 2,
			Children: []*TocEntry{&TocEntry{
				Title:    "Heading 3",
				Level:    3,
				Children: make([]*TocEntry, 0),
			}},
		}, &TocEntry{
			Title:    "Heading 2",
			Level:    2,
			Children: make([]*TocEntry, 0),
		}},
	}, &TocEntry{
		Title: "Heading 1",
		Level: 1,
		Children: []*TocEntry{&TocEntry{
			Title:    "Heading 4",
			Level:    4,
			Children: make([]*TocEntry, 0),
		}, &TocEntry{
			Title: "Heading 2",
			Level: 2,
			Children: []*TocEntry{&TocEntry{
				Title:    "Heading 4",
				Level:    4,
				Children: make([]*TocEntry, 0),
			}},
		}},
	}, &TocEntry{
		Title:    "Heading 1",
		Level:    1,
		Children: make([]*TocEntry, 0),
	}}

	// 	expected0 := `[
	// 		{
	// 				"title": "Heading 1",
	// 				"level": 1,
	// 				"children": [
	// 						{
	// 								"title": "Heading 2",
	// 								"level": 2,
	// 								"children": [
	// 										{
	// 												"title": "Heading 3",
	// 												"level": 3,
	// 												"children": []
	// 										}
	// 								]
	// 						},
	// 						{
	// 								"title": "Heading 2",
	// 								"level": 2,
	// 								"children": []
	// 						}
	// 				]
	// 		},
	// 		{
	// 				"title": "Heading 1",
	// 				"level": 1,
	// 				"children": [
	// 						{
	// 								"title": "Heading 4",
	// 								"level": 4,
	// 								"children": []
	// 						},
	// 						{
	// 								"title": "Heading 2",
	// 								"level": 2,
	// 								"children": [
	// 										{
	// 												"title": "Heading 4",
	// 												"level": 4,
	// 												"children": []
	// 										}
	// 								]
	// 						}
	// 				]
	// 		}
	// ]`

	if reflect.DeepEqual(toc0, expected0) != true {
		t.Error("Table of Contents does not look like expected")
	}
}
