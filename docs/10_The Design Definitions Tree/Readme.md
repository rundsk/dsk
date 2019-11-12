# The Design Definitions Tree

In order for you to document you Design System, DSK expects you to create what we call the “design definitions tree”. It is the manifestation of your Design Systems and the ”database” of DSK.

One of the fundamental ideas in DSK was to use the filesystem as the interface for content creation. This enables _direct manipulation_ of the content and frees us from tedious administration interfaces.

![screenshot](The-Design-Definitions-Tree/Aspects/folder.jpg)

The _design definitions tree_ (DDT for short), is a tree of directories and subdirectories. Each of these directories stands for a _design aspect_ in the hierarchy of your design system, these might be actual components, when you are documenting the user interface, or chapters of your company’s guide into its design culture. In other words: an aspect is a folder that can contain several other aspects and/or several files belonging to this aspect.

→ [Read on about Aspects](./Aspects)

<CodeBlock title="Scheme of a design definitions tree">
<script>
example-ddt
├── AUTHORS.txt                 <- authors database, see "Authors" below
├── Components
│   ├── Button                  <- "Button" design aspect
│   │   └── ...
│   ├── TextField               <- "TextField" design aspect
│   │   ├── Password            <- nested "Password" design aspect
│   │   │   └── readme.md
│   │   ├── api.md              <- document
│   │   ├── exploration.sketch  <- asset
│   │   ├── meta.yml            <- meta data file
│   │   ├── explain.md          <- document
│   │   └── unmask.svg          <- asset
</script>
</CodeBlock>

By default, DSK displays aspects in alphabetic order. You can manually change this order be prepending a number, separated by a dash or underscore to the title of the aspect.

```
example-ddt
├── 01-Visual Design
├── 02-Components
```

<Banner title="Note" type="warning">Hidden directories and files beginning with a dot (.) are ignored. Things you don’t want to be accessed should not be stored inside the DDT.</Banner>

<Banner title="Note" type="warning">Multiple occurrences of the same prepending number in a single directory leads to undefined behaviour.</Banner>
