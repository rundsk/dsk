# Adding Aspect Meta Data (Description, Tags, ...)

To add more information to an aspect, we use a file called `meta.yml`. This 
file holds meta data, like a short description, tags and authors, about an aspect. 
The file uses the easy to write [YAML](https://www.youtube.com/watch?v=W3tQPk8DNbk) 
format.

_Note_: If you prefer to use [JSON](https://www.json.org) as a format,
that is supported too. Just exchange `.yml` for `.json` as the
extension.

An example of a full meta data file looks like this:

```yaml
authors: 
  - christoph@atelierdisko.de
  - marius@atelierdisko.de

description: > 
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.

related:
  - DataEntry/TextField

tags:  
  - priority/1
  - release/0.1
  - progress/empty

version: 1.2.3
```

Possible meta data keys are:

- `authors`: An array of email addresses of the document authors; see below.
- `description`: A single sentence that roughly describes the design aspect.
- `related`: An array of related aspect URLs within DSK.
- `tags`: An array of tags to group related aspects together.
- `version`: A freeform version string.


