# Design System Kit

## Abstract

Using the Design System Kit you quickly define and organize
_design aspects_ into a browsable and live-searchable design system.
Hierarchies between design aspects are established using plain
simple directories. Creating documentation is as easy as adding a
[Markdown](https://guides.github.com/features/mastering-markdown/) formatted
file to a directory inside the _design definitions tree_.

Visit the official [Website](https://rundsk.com), to get started with our [step by step guide](https://rundsk.com/tree/Getting-Started/Step-by-Step).

![DSK promotional image](https://rundsk.com/api/v2/tree/dsk_promo_list.jpg)

## Sponsors

**[Fielmann](https://fielmann.com)** is a German eye-wear company
and market leader in Germany. As our first sponsor, Fielmann has been
extremely supportive in building significant features, i.e. versioning
as well as taking an important role in the bootstrapping phase of DSK.

**[Deutsche Presse-Agentur](https://dpa.com)** is the German
news agency founded in 1949. The agency is primarly sponsoring content
creation features, like individual documentation components, as
well as infrastructure related ones. DSK is used at dpa, to drive one of
its <a href="https://developerdocs.dpa-id.de">public documentation
platforms</a>.

**[Atelier Disko](https://atelierdisko.de)** is a German design and
software studio, developing future oriented solutions for humans and businesses
alike. Atelier Disko provides services around creating and maintaining
top-of-the-industry design systems. Atelier Disko funded the intial development of DSK.

## Quickstart

1. Visit the [GitHub releases page](https://github.com/rundsk/dsk/releases) and download one of the quickstart packages for your operating system. For macOS use `dsk-darwin-amd64.zip`, for Linux use `dsk-linux-amd64.tar.gz`. 

2. The package is an archive that contains the `dsk` executable and an example design system. Double click on the downloaded file to unarchive both. 

3. You start DSK by double clicking on the executable. On first use please follow [these instructions](https://support.apple.com/kb/PH25088) for macOS to skip the developer warning.

4. You should now see DSK starting in a small terminal window, [open the web application in your browser](http://localhost:8080), to browse through the design system.

Too quick? Follow the alternative [Step by Step guide](https://rundsk.com/tree/Getting-Started/Step-by-Step) to get started.

## The Design Definitions Tree

One of the fundamental ideas in DSK was to use the filesystem as the interface for content creation. This enables _direct manipulation_ of the content and frees us from tedious administration interfaces.

![Screenshot of Finder showing a design aspect folder](https://rundsk.com/api/v2/tree/The-Design-Definitions-Tree/Aspects/folder.jpg)

The _design definitions tree_ (DDT for short), is a tree of
directories and subdirectories. Each of these directories stands for
a _design aspect_ in the hierarchy of your design system, these might
be actual components, when you are documenting the user interface, or
chapters of your company's guide into its design culture.

Each directory may hold several files to document these design aspects: a
configuration file to add meta data or supporting _assets_ that can
be downloaded through the frontend.

```
example
├── AUTHORS.txt                 <- authors database, see "Authors" below
├── DataEntry
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
```

Read more about [the design definitions tree](https://rundsk.com/tree/The-Design-Definitions-Tree), and how to add meta data, assets and authors.

## Help

Combining great new ideas with experience will help us create the best possible
features for DSK. Likewise, talking through a bug with a team member will help
us ensure the best possible fix. We're striving to maintain a lean, clean core
and want to avoid introducing patches for symptoms of an underlying flaw.

Found a bug? [Open an issue in our
tracker](https://github.com/rundsk/dsk/issues/new) and label it as a
_bug_.

Have an idea for a new killer feature? [Open an issue in our
tracker](https://github.com/rundsk/dsk/issues/new) and use the
_enhancement_ label.

Just want to say "Thank you" or need help getting started? [Send us a mail](mailto:thankyou@rundsk.com).
