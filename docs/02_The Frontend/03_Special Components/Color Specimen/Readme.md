<!-- # Examples

<ColorSpecimen src="./colors.some"></ColorSpecimen>

## Compact
<ColorSpecimen src="./colors.some" compact="true"></ColorSpecimen>
 -->

# Usage

```
<ColorSpecimen src="./colors.json"></ColorSpecimen>
<ColorSpecimen src="./colors.json" compact="true"></ColorSpecimen>
```

# File format
The Color Specimen component expects a file according to the [Lona Color Defintions Spec](https://github.com/airbnb/Lona/blob/master/docs/file-formats/colors.md).

<CodeBlock title="colors.json">{
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