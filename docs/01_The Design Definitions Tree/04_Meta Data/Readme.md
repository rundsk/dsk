# Meta Data

You can add meta data about an aspect by placing  a special file called `meta.yaml` into the directory of the aspect. This file holds meta data, like a short description, tags and authors, about an aspect. The file uses the easy to write  [YAML](https://www.youtube.com/watch?v=W3tQPk8DNbk)  format.

<Banner title="Note" type="important">If you prefer to use  <a href="https://www.json.org/">JSON</a> as a format, that is supported too. Just exchange <code>.yaml</code> for <code>.json</code> as the extension.</Banner>

An example of a full meta data file looks like this:

<CodeBlock title="meta.yaml">
authors:
  - christoph@atelierdisko.de
  - marius@atelierdisko.de

description: >
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.

related:
  - DataEntry/Dropdown

tags:  
  - priority/1
  - release/0.1
  - progress/draft

version: 1.2.3

custom:
  - synonyms:
    - Input
    - Text Field
  - platform: iOS
</CodeBlock>

Possible meta data keys are:
* `authors`: An array of email addresses of the document authors. See _Authors_ for more information.
* `description`: A single sentence that roughly describes the design aspect.
* `related`: An array of related aspect URLs within DSK.
* `tags`: An array of tags to group related aspects together.
* `version`: A freeform version string.

Starting with DSK version 1.2 (currently in alpha) you can also add custom meta data, under the key `custom`. These items will be displayed alongside the general meta data of the aspect.
