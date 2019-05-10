# Examples

```Component
<Banner title="Nothing special">This is a default banner.</Banner>
<Banner title="Be careful!" type="warning">This is a warning.</Banner>
<Banner title="Oops" type="error">This is an error.</Banner>
<Banner title="Read this" type="important">This is important.</Banner>
```

# Usage

~~~
```Component
<Banner title="Nothing special">This is a default banner.</Banner>
<Banner title="Be careful!" type="warning">This is a warning.</Banner>
<Banner title="Oops" type="error">This is an error.</Banner>
<Banner title="Read this" type="important">This is important.</Banner>
```
~~~

# Properties

Property | Type | Description | Default
---|---|---|---
`title` | `String` | A title to display on top of the banner. | `""`
`type` | `"default"`, `"warning"`, `"error"`, `"important"` | The type of banner to show – this changes the banner’s color. | `default`