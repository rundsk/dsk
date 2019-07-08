# Documents

Aspects are documented by adding [Markdown](https://guides.github.com/features/mastering-markdown/)  formatted documentation files to their directory.

A document file may describe an aspect or give clues how to use a certain component. You can split documentation over several files when you like to. We usually use `readme.md`, `api.md`, `explain.md` or `comments.md`.

<Banner title="Note" type="important"><code>readme.md</code> is in no way treated specially by DSK, but is usually displayed by <a href="https://www.github.com">GitHub</a> as the primary document in the web interface. It is therefore a good idea to use it as the primary document.</Banner>

Using [Markdown](https://guides.github.com/features/mastering-markdown/) it is easy to structure your documents with headlines, format your text and add links, lists, quotes, tables and images. With the built-in fronend you can also use special _Documentation Components_, like banners, do/don’t cards, components playgrounds and color or type specimens to enrich your documentation.

<Banner title="Note" type="important">If you prefer plain HTML documents over Markdown, these are supported too. For this use <code>.html</code> instead of <code>.md</code> as the file extension.</Banner>

Just like aspects, documents can be manually ordered by prefixing them with a number, separated by a dash or underscore to the title of the document.

```
example-ddt
├── 01-Visual Design
├── 02-Components
│   ├── 01-Readme.md
│   └── 02-API.md
```
