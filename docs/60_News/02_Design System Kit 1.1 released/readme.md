# Design System Kit 1.1 released 

With this release we’re significantly extending the search capabilities,
introducing initial Git support and refining the built-in frontend of
[DSK](https://github.com/atelierdisko/dsk).

![](/dsk_promo_list.jpg)

Search is now driven by the go native “[bleve](https://github.com/blevesearch/)”
text indexing library, which enables us to perform advanced analysis on the
searched text and use the best fitting matching algorithms. The result is better
results, even if your query includes small typos or asks for a different form of
the same word.

The – already present – filtering feature has been improved and will now use
prefix and real fuzzy matching under the hood. The full-text search feature is
really brand-new: it helps to find important design aspects even in very large
design systems, where these naturally get lost more easily. We’ve intentionally
chosen to implement search directly within DSK instead of relying on external
services: We are sure that less external dependencies make deployments easier,
especially in corporate environments.

With the introduction of full-text search, we’re dropping support for the
_keywords_ option in the _meta.yaml_ files. The search indexer will “generate”
keywords for you automatically.

The built-in frontend now shows more meta data for each design aspect: author
information, download file sizes, as well as the exact and correct modified
date, which can also be read from Git. We visually cleaned up the frontend in
many places and are using CSS Grid to create the general layout.

The core team likes to take the chance to thank the contributors, who suggested
solutions, tested pre-releases and improved or implemented new features.
Thanks to [Zach Wegrzyniak](https://github.com/wegry/) for contributing to the
search implementation and edge-testing the Git foundations, thanks to [Rodolfo
Marraffa](https://github.com/rodolv1979/) for working on the built-in frontend
with us.

Please see the
[changelog](https://github.com/atelierdisko/dsk/blob/v1.1.0/CHANGELOG.md) for
the complete list of changes. The latest release can be downloaded from the
[releases section](https://github.com/atelierdisko/dsk/releases/tag/v1.1.0) of
our [GitHub project page](https://github.com/atelierdisko/dsk).
