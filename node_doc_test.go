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

// func TestBlockReactComponentInHTML(t *testing.T) {
// 	tmp, _ := ioutil.TempDir("", "tree")
// 	defer os.RemoveAll(tmp)
//
// 	get := func(url string) (bool, *Node, error) {
// 		return false, &Node{}, nil
// 	}
//
// 	node0 := filepath.Join(tmp, "foo")
// 	os.Mkdir(node0, 0777)
//
// 	doc0 := filepath.Join(node0, "readme.html")
// 	raw0 := `
// <script type="text/jsx">
// <CodeBlock title="Example">
//   echo "GREETINGS PROFESSOR FALKEN."
// </CodeBlock>
// </script>
// `
// 	expected0 := `
// <div id="dsk-component-mount-point-123"></div>
// <script>
// React.createElement(CodeBlock, {
//   title: "Example"
// }, "echo 'GREETINGS PROFESSOR FALKEN.'");
// </script>
// `
// 	ioutil.WriteFile(doc0, []byte(raw0), 0666)
//
// 	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
// 	docs, _ := node.Docs()
//
// 	html0, _ := docs[0].HTML("/tree", "foo/bar", get)
//
// 	if string(html0) != expected0 {
// 		t.Errorf("Component markup does not look like expected, got: %s", html0)
// 	}
// }

func TestBlockWebComponentInHTML(t *testing.T) {
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

// func TestBlockWebComponentInMarkdown(t *testing.T) {
// 	tmp, _ := ioutil.TempDir("", "tree")
// 	defer os.RemoveAll(tmp)
//
// 	get := func(url string) (bool, *Node, error) {
// 		return false, &Node{}, nil
// 	}
//
// 	node0 := filepath.Join(tmp, "foo")
// 	os.Mkdir(node0, 0777)
//
// 	doc0 := filepath.Join(node0, "readme.md")
// 	raw0 := `
// <dsk-code-block>
//   echo "GREETINGS PROFESSOR FALKEN."
// </dsk-code-block>
// `
// 	expected0 := `
// <dsk-code-block>
//   echo "GREETINGS PROFESSOR FALKEN."
// </dsk-code-block>
// `
// 	ioutil.WriteFile(doc0, []byte(raw0), 0666)
//
// 	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
// 	docs, _ := node.Docs()
//
// 	html0, _ := docs[0].HTML("/tree", "foo/bar", get)
//
// 	if string(html0) != expected0 {
// 		t.Errorf("Component markup does not look like expected, got: %s", html0)
// 	}
// }

func TestImpliedScriptTagsBlockReactComponentInMarkdown(t *testing.T) {
	raw0 := `
<CodeBlock title="Example">
  echo "GREETINGS PROFESSOR FALKEN."
</CodeBlock>
`
	expected0 := `
<script><CodeBlock title="Example">
  echo "GREETINGS PROFESSOR FALKEN."
</CodeBlock></script>
`
	result0, _ := implyScriptTags([]byte(raw0))
	if string(result0) != expected0 {
		t.Errorf("Failed to imply script tags, got: %s", result0)
	}
}

// func TestBlockReactComponentInMarkdown(t *testing.T) {
// 	tmp, _ := ioutil.TempDir("", "tree")
// 	defer os.RemoveAll(tmp)
//
// 	get := func(url string) (bool, *Node, error) {
// 		return false, &Node{}, nil
// 	}
//
// 	node0 := filepath.Join(tmp, "foo")
// 	os.Mkdir(node0, 0777)
//
// 	doc0 := filepath.Join(node0, "readme.md")
// 	raw0 := `
// <CodeBlock title="Example">
//   echo "GREETINGS PROFESSOR FALKEN."
// </CodeBlock>
// `
// 	expected0 := `
// <div id="dsk-component-mount-point-123"></div>
// <script>
// React.createElement(CodeBlock, {
//   title: "Example"
// }, "echo 'GREETINGS PROFESSOR FALKEN.'");
// </script>
// `
// 	ioutil.WriteFile(doc0, []byte(raw0), 0666)
//
// 	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
// 	docs, _ := node.Docs()
//
// 	html0, _ := docs[0].HTML("/tree", "foo/bar", get)
//
// 	if string(html0) != expected0 {
// 		t.Errorf("Component markup does not look like expected, got: %s", html0)
// 	}
// }

func TestImpliedScriptTagsInlineReactComponentInMarkdown(t *testing.T) {
	raw0 := `
Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
`
	expected0 := `
Yellow and <script><ColorSwatch>green</ColorSwatch></script> are the colors of spring.
`
	result0, _ := implyScriptTags([]byte(raw0))
	if string(result0) != expected0 {
		t.Errorf("Failed to imply script tags, got: %s", result0)
	}
}

// func TestInlineReactComponentInMarkdown(t *testing.T) {
// 	tmp, _ := ioutil.TempDir("", "tree")
// 	defer os.RemoveAll(tmp)
//
// 	get := func(url string) (bool, *Node, error) {
// 		return false, &Node{}, nil
// 	}
//
// 	node0 := filepath.Join(tmp, "foo")
// 	os.Mkdir(node0, 0777)
//
// 	doc0 := filepath.Join(node0, "readme.md")
// 	raw0 := `
// Yellow and <ColorSwatch>green</ColorSwatch> are the colors of spring.
// `
// 	expected0 := `
// <p>
// Yellow and <div class="dsk-component-mount-point-123"></div> are the colors of spring.
// </p>
// <script>
// React.createElement(ColorSwatch, null, "green");
// </script>
// `
//
// 	ioutil.WriteFile(doc0, []byte(raw0), 0666)
//
// 	node := &Node{root: tmp, path: filepath.Join(tmp, "foo")}
// 	docs, _ := node.Docs()
//
// 	html0, _ := docs[0].HTML("/tree", "foo/bar", get)
//
// 	if string(html0) != expected0 {
// 		t.Errorf("Component markup does not look like expected, got: %s", html0)
// 	}
// }
