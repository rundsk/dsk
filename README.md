# Design System Kit

[![Build Status](https://travis-ci.org/atelierdisko/dsk.svg?branch=1.0)](https://travis-ci.org/atelierdisko/dsk)

## Abstract

Using the _Design System Kit_ you quickly define and organize design documentation into a
browsable and live-searchable design system. Hierachies are established using plain simple directories. Documentation is created by just adding a [Markdown](https://guides.github.com/features/mastering-markdown/) formatted file.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk.png)

## Quickstart

Visit the [GitHub releases page](https://github.com/atelierdisko/dsk/releases) and download the binary for your architecture. For macOS use `dsk-darwin-amd64`, for Linux use `dsk-linux-amd64`.

After downloading the binary, you need to make the binary excutable.
```
chmod +x dsk
```

Now run the `dsk` command, giving it the path to the design definitions tree. Double-clicking the command, works fine as well. In that case it'll use the current directory design definitions tree. 
```
./dsk example
```

Finally [open the web application in your browser](http://localhost:8080), to browse through the design system.

## The Design Definitions Tree

The _design definitions tree_ (DDT for short), is a tree of directories and subdirectories. 
Each of these directories stands for a point in the hierarchy of your design system, these
might be actual components, when you are documenting the user interface, or chapters of your
company's guide into its design culture or any other important design aspect.

Each directory may hold files to document the aspects, a configuration file, to add
meta data and supporting _assets_ that can be downloaded through the frontend.

```
example
├── AUTHORS.txt                 <- optional, see "Authors" below
├── DataEntry
│   ├── Button
│   │   ├── exploration.sketch  <- asset
│   │   ├── index.json          <- configuration
│   │   └── readme.md           <- document
│   ├── TextField
│   │   ├── Password
│   │   │   └── readme.md
│   │   ├── api.md              <- another document
│   │   ├── index.json
│   │   ├── readme.md
│   │   └── unmask.svg
```

### Documenting Design Aspects

Directories in the DDT can hold [Markdown](https://guides.github.com/features/mastering-markdown/) formatted documentation files, like `readme.md`, that describe a design aspect or give glues how to use a certain component. Please note
that `readme.md` is in no ways treated specially by DSK, but is usually displayed by GitHub as the primary document of a directory. You can split documentation over several files when you like to. i.e. We usually use `api.md`,
`explain.md` or `comments.md`.

### Design Configuration File

Directories in the DDT may also hold an `index.json` file. This file and any of its
configuration settings are entirely optional. 

Using the configuration file, we can add meta data to the design aspect (i.e. tags)
and improve the search experience in the interface. 

An example of a full configuration looks like this:

```json
{
    "authors": ["christoph@atelierdisko.de", "marius@atelierdisko.de"],
    "description": "This is a very very very fancy component.",
    "keywords": ["typography", "font", "type"],
    "tags": ["fancy", "very"],
    "version": "1.2.3"
}
```

Possible configuration options are:

- `authors`: An array of email addresses of the document authors; see below.
- `description`: A single sentence that roughly describes the design component.
- `keywords`: An array of terms that are searched in addition to `tags`.
- `tags`: An array of tags to group related design components together.
- `version`: A freeform version string.

### Authors

Each directory inside the tree may be _authored_ by one or multiple
persons. To assign yourself, use the `authors` option in `index.json`
the design configuration file.

To enable automatic full names for each author, create an
`AUTHORS.txt` file inside the root of the DDT. Each line of the
file lists an author's full name and her/his email address in angle
brackets.

```text
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <marius@atelierdisko.de>
```

# Architecture

Architecture-wise DSK is split into a backend and frontend. The backend implemented 
in Go takes care of understanding the defintions tree and provides a REST API for the
frontend, usually implemented in JavaScript. 

Frontends are pluggable and the decoupled desing allows you to create indvidually branded frontends. 
These are entirely free in their implementation, they must adher to only a minimal set
of rules.

The frontend and backend and are than later compiled together into a single binary, making
it usuable as a publicly hosted web application or a locally running design tool.

## Building your own Frontend 

The following sections describe everything you need to know when building your own frontend
and bundle it with the dsk binary. By default dsk uses a builtin minimal frontend. The default frontend
is a good starting point when developing your own. It can be found in the `frontend` directory of
this project.

### Available API Endpoints

The backend provides the following API endpoints. JSON responses use the
[JSend](https://labs.omniti.com/labs/jsend) format.

`GET /api/v1/tree`,
get the full design definitions tree as a nested tree of nodes.

`GET /api/v1/tree/{path}`,
get information about a single node specified by `path`.

`GET /api/v1/tree/{path}/{asset}`,
requests a node's asset, `{asset}` is a single filename.

`GET /api/v1/search?q={query}`,
performs a search over the design definitions tree and returns
a flat list of matched node URLs.

### Designing the URL Schema

Your frontend and its subdirectories will be mounted directly at the root path
`/`. Requests to anything under `/api` are routed to the backend, anything else
is routed into your application in `index.html`. 

Relative asset source paths inside the markdown files will be made
absolute to allow you displaying the document contents wherever you
like to.

### Baking

To _bake_ your frontend into DSK, install the the development tools as described in the _Development_ section frist. 
After doing so, you create your custom dsk build by running the following command.

```
$ FRONTEND=/my/frontend make dist
```

Frontends created with [create react app](https://github.com/facebook/create-react-app) should instead follow these couple of simple steps. 

```
$ cd /my/frontend
$ npm run build
$ cd $(go env GOPATH)/github.com/atelierdisko/dsk
$ FRONTEND=/my/frontend/build make dist
```

## Development

[Go](https://golang.org/) version 1.9 or later is needed for developing and
testing the application. 

If you're setting up a Go environment from scratch, add Go's `bin` directory to
your `PATH`, so that go binaries like `go-bindata` can be found. In `.profile`
add the following line.
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

To run the unit tests use `make test`.

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.
