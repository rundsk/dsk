# Example

```Component
<AnnotatedImage src="image.png" annotations="annotations.json"></AnnotatedImage>
```

# Usage

~~~
```Component
<AnnotatedImage src="image.png" annotations="annotations.json"></AnnotatedImage>
```
~~~

## Annotation Format

```Component
<CodeBlock title="annotations.json" language="json">{
  "annnotations": [
    {
      "x": "30%",
      "y": "10%",
      "description": "Use a clear label"
    },
    {
      "x": "45%",
      "y": "36%",
      "description": "Pick a color with enough contrast",
      "offsetX": "50px"
    }
  ],
  "annotationColor": "#FF00FF"
}</CodeBlock>
```