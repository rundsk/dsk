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

The _design definitions tree_ (DDT for short), is a directory containing other nested directories. Each one most often at least contains a `readme.md` file. This is simply a Markdown formatted text file that. Other Markdown files are supported too like `api.md` and `comments.md`. There is the configuration file `index.json` which we describe in the following section. Any other file will be considered an _Asset_ that is it can be embedded inside the Markdow files or downloaded throug the frontend.

```
example
├── AUTHORS.txt                 <- optional, see "Owners & Authors" below
├── DataEntry
│   ├── Button
│   │   ├── exploration.sketch  <- Asset
│   │   ├── index.json          <- Configuration
│   │   └── readme.md           <- Main markdown file
│   ├── TextField
│   │   ├── Password
│   │   │   └── readme.md
│   │   ├── api.md
│   │   ├── index.json
│   │   ├── readme.md
│   │   └── unmask.svg
```

### Design Configuration File

Each tree directory may hold an `index.json` file. The file and any of its
configuration settings are entirely optional.

Using the configuration file, we can add meta data to the design (i.e. keywords)
to improve the search experience in the interface. 

An example of a full configuration looks like this:

```json
{
    "description": "This is a very very very fancy component.",
    "keywords": ["fancy", "very"]
    "owners": ["christoph@atelierdisko.de", "marius@atelierdisko.de"]
}
```

Possible configuration options are:

- `description`: A single sentence that roughly describes the design component.
- `keywords`: An array of keywords to group related design components together.
- `owners`: An array of email addresses of the document owners; see below.
- `version`: A freeform version string.

### Authors & Owners

Each directory inside the tree may be _owned_ by one or multiple
authors. To assign youreself, use the `owners` option in `index.json`
the design configuration file.

To enable automatic full names for each owner, create an `AUTHORS.txt`
file inside the root of the design definitions tree. Each line of the
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

`GET /api/v1/tree?q={query}`,
get a filtered by `query` view onto the design defintions tree.

`GET /api/v1/tree/{path}`,
get information about a single node specified by `path`.

`GET /api/v1/tree/{path}/{asset}`,
requests a node's asset, `{asset}` is a single filename.

`GET /api/v1/search?q={query}`,
full text search over the design definitions tree.

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
