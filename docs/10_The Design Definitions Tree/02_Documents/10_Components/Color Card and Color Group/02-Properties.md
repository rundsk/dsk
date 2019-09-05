# Properties

## Color

Property | Type | Description | Default
---|---|---|---
`children` | `String` | The name of the color | [Required]
`color` | `String` | The color value, in CSS format. | [Required]
`id` | `String` | A code-friendly id of the color. | `""`
`comment` | `String` | An optional description of the color, explaining contextual information, such as how it should be used. | `""`
`compact` | `Bool` | Whether the color card should be displayed in a compact way. | `false`

## Color Group

Property | Type | Description | Default
---|---|---|---
`children` | `[ColorCard]` | One or more color cards. | `[]`
`src` | `String` | Path to a color specification file. | `""`
`compact` | `Bool` | Whether the color cards should be displayed in a compact way. | `false`

The Color Group component expects a file according to the [Lona Color Defintions Spec](https://github.com/airbnb/Lona/blob/master/docs/file-formats/colors.md).