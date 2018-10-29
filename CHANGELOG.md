# Changelog

## [1.1.0] Unreleased

- Add HTTP APIv2
- Introduce full search in APIv2, #44
- New and improved filter search in APIv2: 
  - now uses prefix matching instead of haystack/needle,
  - slightly more immune to typos by using fuzzy matching,
  - uses analyzers as full search does,
  - supports `wide` mode, 
  - API responses now use `nodes` key instead of `urls` to return an
    array of so-called _RefNodes_ with title and URL for each node. This
    makes that part of the response uniform when compared to other API
    responses.
  - keywords are not searched anymore.
- Add initial Git support and read modified dates from it, if possible
- Replace go-bindata build time dependency with vfsgen, fixes #40 and #49
- Require Go v1.11, drop support for Go v1.9 and v1.10
- Improve transliteration when creating node slugs
- Normalize strings when they are read from the filesystem, fixes #48.
- Deprecate `keywords` in meta data

## [1.0.2] 2018-09-06

- Add support for Go v1.11

## [1.0.1] 2018-06-27

- Normalize strings when they are read from the filesystem, fixes #48.

## [1.0.0] 2018-05-15
