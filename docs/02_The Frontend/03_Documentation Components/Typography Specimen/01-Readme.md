<Banner title="Version Feature">
  Documentation Components are available since version 1.2.
</Banner>

# Example

<TypographySpecimen src="./typography.json"></TypographySpecimen>

# Usage

```
<TypographySpecimen src="./typography.json"></TypographySpecimen>
```

# File format
The Typography Specimen component expects a file according to the [Lona Text Style Spec](https://github.com/airbnb/Lona/blob/master/docs/file-formats/text-styles.md).

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