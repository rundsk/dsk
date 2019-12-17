# Changelog

## 1.3.0 Unreleased

- We've added support for multiple versions of a DDT
- We've moved to a new GitHub organization
- The docs have been extracted and moved to https://github.com/rundsk/website
- The JavaScript SDK has been extracted and moved to https://github.com/rundsk/js-sdk
- The example design system has been extracted and moved to https://github.com/rundsk/example-design-system
- The backend has been refactored and modularized
- The built-in frontend now supports captions for images, i.e. `<Image caption="Hello World">`
- The annotations for playgrounds have been improved
- We've removed support for `go get` without additional initialization as this required us to
  keep a comparitvely big go file around where assets had been inlined
- We've fixed a bug where assets with an order number could not be loaded
- We now support HTML comments inside Markdown documents
- We've moved from an internal Concourse pipeline to GitHub actions, which will run our test
  and deploy the website. We are now able to automatically build and test packages for macOS.
- API responses for node assets that are images, now carry its dimensions in pixels.
- JSON and YAML assets can now be converted into the other type, just exchange the extension
  to either `.json` or `.yaml`, when forming the URL to request the asset. To render the `foo.yaml`
  asset as JSON use `/api/v1/tree/foo.json` instead of `/api/v1/tree/foo.yaml`
- To control resource sharing we've introduced the `-allow-origin` flag, which
  needs to be provided when starting DSK, if you like to allow origins from which
  browsers can access API resources. A common scenario for this is when you have
  multiple DSKs running and the frontend of DSK (a) wants to access the API of DSK
  (b), i.e. for cross references.
- `Link` does now support providing a `target`. 

## 1.2.0

- The frontend has been rewritten as a React-App, and redesigned
  from scratch. It surfaces important features, like full text search and
  just looks great.
- We've introduced documentation components, i.e. `<Banner>`, `<CodeBlock>`, or `<DoDont>` 
  to help you creating   top class design documentation. This feature is currently limited 
  to Markdown documents.
- We now support freeform data under the `custom` key in `meta.yaml` (or `meta.json`). It 
  is made available through the API. The built-in frontend has been enhanced to display 
  the freeform meta data nicely. #62, #71, #72
- Certain aspects can now be configured through an optional configuration file.
- The window title is now formatted in a more search engine/human friendly way.
  When an aspect has a description we use that to set a meta description.
- We now clean up Markdown documents better before they are indexed for search.
- Search results now return hit fragments, so highlighting can be implemented in the frontend.
- We unfiltered assets, so that these can be freely inlcuded or made available for
  download in the frontend. It's now up to the frontend to decided whether an asset
  should be made available for download. The next version of the API will not 
  include `downloads` on node results any more. We recommend to use the new `assets` 
  instead. 
- The DDT is not a good place to store secrets and we want to be more clear about that. We've
  removed special ignore rules and do not exclude aspects  prefixed with  `_`, `x-` or `x_`. 
  anymore. This clarifies a possible misunderstanding. Things you don't want to be accessed, 
  should not be stored inside the DDT.
- Each document now makes an array of (top-level) components available
  through the API. We plan to expand the information that 
  is returned here for the next release.
- We fixed an issue where the backend didn't correctly respond to HTTP
  requests. The backend now responds to missing API resources with a
  404, this allows to use HEAD request to test for the existence of i.e.
  a node.
- Our JavaScript SDK `frontend/js/dsk` is now a package, that we publish on npm.
- The JavaScript API client from the SDK, now supports pinging and    
  checking for the existence of a node or asset. In total three new     
  methods were added: `has()`, `config()` and `ping()`.                 
- The document transform from the JavaScript SDK, is used to parse documents
  received from the backend and replaces occurences of components, with 
  actual component instances. The transformer has cleaned up and modified 
  to work with the new documentation components.
- Many of you wish to run DSK inside a docker container. We now prebuild and ship official 
  container imagee and make them available on
  [on docker hub](https://cloud.docker.com/u/atelierdisko/repository/registry-1.docker.io/rundsk/dsk).
- Support for multiple languages has been removed, this unfortunately
  didn't work out the way we expected to. We've also removed the `lang`
  command line in favor of providing the language of the DDT via the new
  optional configuration file. In the language is a property of the DDT
  not the program.
- We've also removed support for using wide index with filter search,
  all filter queries will use the narrow index.
- The `frontend` flag has been added, so the built-in frontend can be switched out without
  compiling it in.

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
