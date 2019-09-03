<Banner title="Version Feature">
  Documentation Components are available since version 1.2.
</Banner>

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

For all configruation options see [Properties](?t=properties).

## Annotations

With annotations you can highlight specific points on the playground and add a comment.

<Playground annotations="annotations.json">{Component}</Playground>

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