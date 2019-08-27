<Banner title="Experimental Feature" type="warning">Documentation Components are a new feature and will be introduced with version 1.2, which is currently in alpha.</Banner>

# Examples

## Basic

<CodeBlock>
alert('Hello World!');
</CodeBlock>

```
<CodeBlock>
alert('Hello World!');
</CodeBlock>
```

## Documenting Components

<CodeBlock>
<Color color="#001dff">Blue</Color>
</CodeBlock>

```
<CodeBlock>
<Color color="#001dff">Blue</Color>
</CodeBlock>
```


## With Title

<CodeBlock title="fib.js">
function fib(n) {
  return n < 2 ? n : fib(n - 1) + fib(n - 2);
}	
</CodeBlock>

```
<CodeBlock title="fib.js">
function fib(n) {
  return n < 2 ? n : fib(n - 1) + fib(n - 2);
}	
</CodeBlock>
```

## Retrieving the Source from a File

<CodeBlock src="./colors.json"></CodeBlock>

```
<CodeBlock src="./colors.json"></CodeBlock>
```

<CodeBlock src="https://gist.githubusercontent.com/adamwathan/b271d1a34f5b37b1a2ad2e844c86b329/raw/d7321013ab5a5b4469b306f3e3a73ee6c0226100/tdd-books.md"></CodeBlock>

```
<CodeBlock src="https://gist.githubusercontent.com/adamwathan/b271d1a34f5b37b1a2ad2e844c86b329/raw/d7321013ab5a5b4469b306f3e3a73ee6c0226100/tdd-books.md"></CodeBlock>
```
