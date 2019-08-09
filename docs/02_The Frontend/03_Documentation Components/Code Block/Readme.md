<Banner title="Experimental Feature" type="warning">Documentation Components are a new feature and will be introduced with version 1.2, which is currently in alpha.</Banner>

# Examples

<CodeBlock title="colors.json">{code from colors.json}</CodeBlock>

<CodeBlock title="Example">
{code}
</CodeBlock>

<CodeBlock title="Example">
  <Color color="#FF0000">Red</Color>
</CodeBlock>

<CodeBlock title="Example">
  <div>test</div>
</CodeBlock>

<Playground>
  <Color color="#FF0000">Red</Color>
</Playground>


# Usage

```
<CodeBlock src="./colors.json"></CodeBlock>
<CodeBlock title="Example">{code}</CodeBlock>
```

# Properties

The content can either be passed as children or be loaded from a file.

Property | Type | Description | Default
---|---|---|---
`src` | `String` | Path to a code file to display. | `null`
`title` | `String` | Title of the code block. | `""`