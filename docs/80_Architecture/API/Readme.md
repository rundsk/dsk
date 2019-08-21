# API

## Versioned API

The API version is a single integer that gets incremented with each version
release. We provide a stability guarantee for our version APIs: Backwards
compatibility breaking changes are only allowed in new major versions of the
API. All other changes are simply additions, optimizations or bug fixes.

## How URLs are Constructed

All endpoint URLs are constructed using the `/api` prefix, the version segment and
the endpoint URL fragment from the tables below, i.e. `/api/v1/hello`.

## Version 2

Version 2 of the API first appeared with DSK Version 1.1.

| URL                             | Response  | Description                       |
|---------------------------------|-----------|-----------------------------------|
| `/hello`                        | JSON      | Returns the version and a friendly greeting |
| `/tree`                         | JSON      | Get the full design definitions tree as a nested tree of nodes |
| `/tree/{path}`                  | JSON      | Get information about a single node specified by `{path}` |
| `/tree/{path}/{asset}`          | data      | Requests a node's asset, `{asset}` is a single filename |
| `/search?q={query}`             | JSON      | Performs a full text search |
| `/filter?q={query}`             | JSON      | Filters nodes by given query |
| `/messages`                     | WebSocket | For receiving messages, i.e. whenever the tree changes |

### Filtering & Searching

#### But how do `filter` and `search` differ? 

`filter` by default matches on all the node's "visible" attributes, it always
returns all matched results without pagination and weighting. Each result is
unique in the set and the total number of results and cannot exceed the number
of total nodes in the tree. We created `filter` for the purpose to drive the
left hand tree navigation of the built-in frontend.

`search` does a lot more work before actually matching, by analyzing context
and content thoroughly. We imagine that even information about actual search
behavior can later be included and used to improve search. `search` is used
to allow users not familar with the design system to perform research and
associative browsing. _Hits_ are ranked by relevance and a node may appear
multiple times in the result set.

|                   | _Fulltext Search_ | _Filtering_ | _Finding_ |
| ----------------- | ----------------- | ----------- | --------- |
| As seen in… | Google | Tree-Nav | Sublime |
| Purpose | research | navigate | jump |
| Possible number of results | 1 million | 20-30 | 1  |
| Result type | hit | result | finding |
| | you don't know your goal | set always contains goal | get to your goal |
| | as good as possible | as quick as possible | as exact as possible |
| | aggregated | realtime | realtime
| Order | weighted by relevance | no weighting | weighted by relevance |
| Match | matches on content and context | matches on _visible_ data | matches on single field | 
| Typical Query Matching | term, phrase after thorough analyzation phase | prefix, keyword, fuzzy with small distance | fuzzy |
| | not too greedy by character | some error tolerance | greedy |

## Version 1

Version 1 of the API first appeared with DSK Version 1.0.

| URL                             | Response  | Description                       |
|---------------------------------|-----------|-----------------------------------|
| `/hello`                        | JSON      | Returns the version and a friendly greeting. |
| `/tree`                         | JSON      | Get the full design definitions tree as a nested tree of nodes. |
| `/tree/{path}`                  | JSON      | Get information about a single node specified by `{path}`. |
| `/tree/{path}/{asset}`          | data      | Requests a node's asset, `{asset}` is a single filename. |
| `/filter?q={query}`             | JSON      | Performs a narrow restricted fuzzy search. |
| `/messages`                     | WebSocket | For receiving messages, i.e. whenever the tree changes. |
