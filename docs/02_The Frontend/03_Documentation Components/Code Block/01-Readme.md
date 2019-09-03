<Banner title="Version Feature">
  Documentation Components are available since version 1.2.
</Banner>

# Examples

## Basic Usage

Usually you'll be using fenced code blocks (\`\`\`) when authoring Markdown. 
These get automatically converted to a `<CodeBlock>`.

```
alert('Hello World!');
```

## Necessity of using `<script>` tags

Whenever the content you place in a `<CodeBlock>`,
includes characters that are used by HTML, you must
wrap the content inside the `<CodeBlock>` in
`<script>` tags, stating that this is literal data.

<CodeBlock>
<script>
alert('Hello World!');
</script>
</CodeBlock>

```
<CodeBlock>
<script>
alert('Hello World!');
</script>
</CodeBlock>
```

## Documenting Components

As any other code, components can be documented using the `<CodeBlock>`, too.

<CodeBlock>
<script>
<Color color="#001dff">Blue</Color>
</script>
</CodeBlock>

```
<CodeBlock>
<script>
<Color color="#001dff">Blue</Color>
</script>
</CodeBlock>
```

## Adding a Title

Some code snippets benefit greatly from an added title. Here we want to show
what the contents of a file called `fib.js` look like.

<CodeBlock title="fib.js">
<script>
function fib(n) {
  return n < 2 ? n : fib(n - 1) + fib(n - 2);
}	
</script>
</CodeBlock>

```
<CodeBlock title="fib.js">
<script>
function fib(n) {
  return n < 2 ? n : fib(n - 1) + fib(n - 2);
}	
</script>
</CodeBlock>
```
