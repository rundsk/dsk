# Design System Kit (DSK)

## Abstract

Using Design System Kit (DSK) you quickly organize components into a
browsable and live-searchable component library.

Hierachies between components are established using plain simple directories.
Creating documentation is as easy as adding a Markdown formatted file to a
directory inside the _design definitions tree_.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk.png)

## Status

DSK is currently in development and **should not yet be considered for general production use**. 
Please see the _Development_ section for how to build the tool from source. We encourage contributions in form of code, design ideas or testing.

## Quickstart

Visit [the releases page](https://github.com/atelierdisko/dsk/releases) and download the binary 
from GitHub, i.e. with cURL.
```
curl -L https://github.com/atelierdisko/dsk/releases/download/v0.6.0-alpha/dsk-darwin-amd64 -o dsk
```

Before you can run the downloaded binary, you must make it executable.
```
chmod +x dsk
```

Now run the `dsk` command pointing it to the directory that contains your design definitions tree.
```
./dsk example
```

Finally [open the Web Application in your browser](http://localhost:8080).

## The Design Definitions Tree

```
example
├── DataEntry
│   ├── Button
│   │   ├── exploration.sketch
│   │   ├── index.json
│   │   └── readme.md
│   ├── TextField
│   │   ├── Password
│   │   │   └── readme.md
│   │   ├── api.md
│   │   ├── index.json
│   │   ├── readme.md
│   │   └── unmask.svg
```

### Component Configuration File

Each component directory may hold an `index.json` file. The file and any of its
configuration settings are entirely optional.

Using the configuration file, we can add meta data to the component (i.e. keywords)
to improve the search experience in the interface. 

An example of a full component configuration looks like this:

```json
{
    "description": "This is a very very very fancy component.",
    "keywords": ["fancy", "very"]
}
```

Possible configuration options are:

- `description`: A single sentence that roughly describes the component.
- `keywords`: An array of keywords to group related components together.

# Architecture

Architecture-wise DSK is split into backend and frontend. The backend implemented 
in Go takes care of understanding the defintions tree and provides a REST API for
the frontend, usually implemented in JavaScript.

The decoupled design allows you to create indvidually branded frontends, which
are entirely free in their implementation, they must adher to only a minimal set
of rules.

The frontend and backend and are compiled together into a single binary, making
usuable as a publicly hosted web application or a locally running design tool.

## Building your own Frontend 

The following sections describe everything you need to build your own frontend
and bundle it with the dsk binary. By default dsk uses the _default frontend_,
which - for inspirational purposes - can be found in the `frontend` directory of
this project.

### Available API Endpoints

The backend provides the following API endpoints. All endpoints return JSON
using [JSend](https://labs.omniti.com/labs/jsend) as the general response
format.

`GET /api/v1/tree`
Get the full design definitions tree as a nested tree of nodes.

`GET /api/v1/tree/{path}`
Get information about a single node specified by `path`.

`GET /api/v1/search?q={query}`
Full text search over the design definitions tree.

### Designing the URL Schema

Your frontend and its subdirectories will be mounted directly at the
root path `/`. You must ensure the frontend doesn't include directories which collide 
with reserved backend paths (currently just `api` is reserved).

So relative media references work in the rendered HTML the frontend should:

1. use the path of the node, i.e. `/Button`, as the canonical URL,
   to display node information. Assets for the `Button` node are served
   _by the backend_ under i.e. `/Button/example.png`.

2. redirect requests for i.e. `/Button` to `/Button/`

A build created by create react app's `npm run build` is a valid frontend:
```
.
├── index.html
└── static
    ├── css
    │   ├── main.41064805.css
    ├── js
    │   ├── main.5f57358c.js
    └── media
        └── exampleImage.3780b1a4.png
```

### Building DSK with your Frontend

Please install the development tools as described in the _Development_ section,
than use the following command to compile a dsk binary with your frontend.

```
$ FRONTEND=/path/to/my/frontend make dist
```

## Development

[Go](https://golang.org/) version 1.9 or later is needed for developing and
testing the application. 

If you're setting up a Go environment from scratch, add
Go's `bin` directory to your `PATH`, so that go binaries like `go-bindata` can
be found. In `.profile` add the following line.
```
export PATH=$PATH:$(go env GOPATH)/bin
```

The `make dev` command assumes your test design system definitions are below a
directory called `_test`. The vendored dependencies are simple Git submodules 
and can be managed manually or with [Manul](https://github.com/kovetskiy/manul).

```
$ go get github.com/kovetskiy/manul
$ go get github.com/twitter/go-bindata/...
$ go get github.com/atelierdisko/dsk
$ cd $(go env GOPATH)/src/github.com/atelierdisko/dsk
$ make dev
```

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.

