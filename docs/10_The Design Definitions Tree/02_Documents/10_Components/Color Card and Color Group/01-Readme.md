<Banner title="Version Feature">
  Documentation components are available since version 1.2.
</Banner>

Moving the mouse over a color card reveals accessibility information about the contrast ratio of black and white to the color. Clicking a color card copies the color’s value.

# Examples

## Single Color

<Color color="#001dff">Blue</Color>

```
<Color color="#001dff">Blue</Color>
```

### Compact

<Color color="#001dff" compact="true">Blue</Color>

```
<Color color="#001dff" compact="true">Blue</Color>
```

## Color Group

<ColorGroup>
  <Color color="#001dff">Blue</Color>
  <Color color="#FFE874" comment="A juice shade of yellow!">Yellow</Color>
</ColorGroup>

```
<ColorGroup>
  <Color color="#001dff">Blue</Color>
  <Color color="#FFE874" comment="A juice shade of yellow!">Yellow</Color>
</ColorGroup>
```

### Compact

<ColorGroup compact="true">
  <Color color="#001dff">Blue</Color>
  <Color color="#FFE874">Yellow</Color>
</ColorGroup>

```
<ColorGroup compact="true">
  <Color color="#001dff">Blue</Color>
  <Color color="#FFE874" comment="A juice shade of yellow!">Yellow</Color>
</ColorGroup>
```

## Color Group from JSON

<ColorGroup src="colors.json"></ColorGroup>

```
<ColorGroup src="colors.json"></ColorGroup>
```

### File format

The Color Group component expects a file according to the [Lona Color Defintions Spec](https://github.com/airbnb/Lona/blob/master/docs/file-formats/colors.md).

<CodeBlock title="colors.json" src="colors.json"></CodeBlock>
