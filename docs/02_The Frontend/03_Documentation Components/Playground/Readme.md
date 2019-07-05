# Examples

## Basic Usage

<Playground>{Component}</Playground>

```
<Playground>{Component}</Playground>
```

## Configured

<Playground background="pinstripes" backgroundColor="#FED28C">{Component}</Playground>

```
<Playground background="pinstripes" backgroundColor="#FED28C">{Component}</Playground>
```

For all configruation options see [Properties](#).

## Annotations

With annotations you can highlight specific points on the playground and add a comment.

<Playground annotations="annotations.json"><FigmaEmbed document="Ppu4fKoeiDXCGMB5XgvZefHc" frame="TagComponent" token="11435-1dd12ee1-db3f-4c56-8e3f-85840e1db2d2"></FigmaEmbed></Playground>

```
<Playground annotations="annotations.json">{Component}</Playground>
```

The annotations specification file has to be formated like this (`annotationColor` is optional):

<CodeBlock title="annotations.json">{
  "annnotations": [
    {
      "x": "36%",
      "y": "0%",
      "description": "Use a clear label"
    },
    {
      "x": "36%",
      "y": "61%",
      "description": "Pick a color with enough contrast",
    }
  ],
  "annotationColor": "#EE645D"
}</CodeBlock>

# Properties

Property | Type | Description | Default
---|---|---|---
`background` | `"dotgrid"`, `"checkerboard"`, `"pinstripes"`, `"plain"` | The pattern to use for the background of the playground. | `"dotgrid"`
`backgroundColor` | `CSS Color String (hex, rgba)` | The background color of the playground. | `#F2F6F7`
`annotations` | `string` | Path to an annotations specification file. | `""`
