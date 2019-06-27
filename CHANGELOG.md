# Changelog

## 1.2.0 Unreleased

- The frontend has been rewritten as a React-App, and redesigned
  from scratch.
- `frontend/js/dsk` is now a package
- Introduce `frontend` flag, so the built-in frontend can be switched out without
  compiling it in.
- Before we index markdown documents for search, they are now
  better cleaned up.
- Implement search hit fragments
- Introduce unfiltered assets, that can be freely inlcuded or made available for
  download in the frontends: stop filtering downloads overly strict, and expose
  all files that are not node documents or meta files as assets. Add `assets` to
  API responses for nodes and deprecate `downloads`, it will be removed in APIv3.
- Remove ignore of directories beginning with an underscore (`_`), `x-` or `x_`.
  This clarifies a possible misunderstanding. Things you don't want to be accessed, 
  should not be stored inside the DDT.
- Official prebuilt docker container images are now available
  [on docker hub](https://cloud.docker.com/u/atelierdisko/repository/registry-1.docker.io/atelierdisko/dsk).
- We now support freeform data under the `custom` key in `meta.yaml` (or `meta.json`). It 
  is made available through the API. The built-in frontend has been enhanced to display 
  the freeform meta data nicely. #62, #71, #72

## 1.1.1

- Fix possible data race in repository lookup table, #58
- Fix issue where "Source" wasn't clickable in the built-in frontend, #66
- Fix incorrect date shown on downloads in the built-in frontend, #57
- Keep original query and fragment when making node URLs absolute, #65

## 1.1.0

The first minor release following the release of 1.0.0, featuring a 
brandnew search build on top of the go native 
[bleve](https://github.com/blevesearch/bleve) 
and laying foundation by adding initial Git support.

- Add HTTP APIv2
- Introduce full search in APIv2, #44
- New and improved filter search in APIv2: 
  - now uses prefix matching instead of haystack/needle,
  - slightly more immune to typos by using fuzzy matching,
  - uses analyzers as full search does,
  - supports `wide` mode, 
  - API responses now use keys named `nodes` instead of `urls` to return an
    array of so-called _RefNodes_ with title and URL for each node. This
    makes that part of the response uniform when compared to other API
    responses.
  - keywords are not searched anymore.
- Add initial Git support and read modified dates from it, if possible
- Replace go-bindata build time dependency with vfsgen, fixes #40 and #49
- Require Go v1.11, drop support for Go v1.9 and v1.10
- Use go modules in favor of of go dep
- Improve transliteration when creating node slugs
- Normalize strings when they are read from the filesystem, fixes #48.
- Deprecate `keywords` in meta data
- Rewrite built-in frontend

Thanks to [Zach Wegrzyniak](https://github.com/wegry/) for contributing to 
the search implementation and edge-testing the Git foundations.

## 1.0.2 2018-09-06

- Add support for Go v1.11

## 1.0.1 2018-06-27

- Normalize strings when they are read from the filesystem, fixes #48.

## 1.0.0 2018-05-15

This is the first stable release of DSK, which we introduce in the
[release announcement post](https://atelierdisko.de/journal/post-167-dsk).
