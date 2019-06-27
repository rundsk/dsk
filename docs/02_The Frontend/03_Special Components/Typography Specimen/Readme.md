<!-- # Examples

```Component
<TypographySpecimen src="./typography.some"></TypographySpecimen>
```

## Compact
```Component
<TypographySpecimen src="./typography.some" compact="true"></TypographySpecimen>
``` -->

# Usage

~~~
```Component
<TypographySpecimen src="./typography.json"></TypographySpecimen>
<TypographySpecimen src="./typography.json" compact="true"></TypographySpecimen>
```
~~~

# File format
The Typography Specimen component expects a file according to the [Lona Text Style Spec](https://github.com/airbnb/Lona/blob/master/docs/file-formats/text-styles.md).

```Component
<CodeBlock title="typography.json">{
  "colors": [
    {
      "name": "Blue",
      "id": "blue",
      "value": "#001dff",
      "comment": "Our primary color"
    },
    {
      "name": "White",
      "id": "white",
      "value": "#ffffff"
    },
    {
      "name": "Green",
      "id": "green",
      "value": "#52d0af"
    },
    {
      "name": "Teal",
      "id": "teal",
      "value": "#0091ab"
    }
  ]
}</CodeBlock>
```

# Properties

Property | Type | Description | Default
---|---|---|---
`src` | `String` | Path to a color specification file. | [Required]
`compact` | `Bool` | Whether the color specimen should be displayed in a compact way. | `false`