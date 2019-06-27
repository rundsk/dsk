You can configure the frontend by placing a file called `frontendConfiguration.json` in the root of your Design Definiftions Tree.

This configuration file could look like this:

```Component
<CodeBlock title="frontendConfiguration.json">{
  "organisation": "ACME Crop",
  "tags": [
    {
      "name": "production",
      "color": "#8DE381"
    }, 
    {
      "name": "deprecated",
      "color": "#ED6666"
    },
    {
      "name": "progress",
      "color": "#0091AB"
    }
  ]
}</CodeBlock>
```

Property | Type | Description
---|---|---|
`organisation` | `String` | The name of the organisation that this Design System is for. Will be displayed in the top left corner of the UI.
`tags` | `[TagConfiguration]` | An array of configuration objects for specific tags. Allows you to display certain tags in custom colors.

## Tags

<figure>
  <img src="tags-example@2x.png">
  <figcaption>The result of the example configuration above</figcaption>
</figure>

Configuring tags allows you to define which color is used for a specific tag.

Property | Type | Description
---|---|---|
`name` | `String` | The color will be used if a tag **contains** this string.
`color` | `CSS Color String (hex, rgba)` | The color to use for the tag.

```Component
<Banner title="Tip">If you want to use colors that blend in with the DSK frontend you can also use one of the following values: <code>"var(--color-teal)"</code>, <code>"var(--color-yellow)"</code>, <code>"var(--color-orange)"</code>, <code>"var(--color-red)"</code>.</Banner>
```