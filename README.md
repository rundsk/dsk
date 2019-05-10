# Design System Kit

[![Build Status](https://travis-ci.org/atelierdisko/dsk.svg?branch=master)](https://travis-ci.org/atelierdisko/dsk)

## Abstract

Using the Design System Kit you quickly define and organize
_design aspects_ into a browsable and live-searchable design system.
Hierarchies between design aspects are established using plain
simple directories. Creating documentation is as easy as adding a
[Markdown](https://guides.github.com/features/mastering-markdown/) formatted
file to a directory inside the _design definitions tree_.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk.png?v=3)

## Quickstart

1. Visit the [GitHub releases page](https://github.com/atelierdisko/dsk/releases) and download one of the quickstart packages for your operating system. For macOS use `dsk-darwin-amd64.zip`, for Linux use `dsk-linux-amd64.tar.gz`. 

2. The package is an archive that contains the `dsk` executable and an example design system. Double click on the downloaded file to unarchive both. 

3. You start DSK by double clicking on the executable. On first use please follow [these instructions](https://support.apple.com/kb/PH25088) for macOS to skip the developer warning.

4. You should now see DSK starting in a small terminal window, [open the web application in your browser](http://localhost:8080), to browse through the design system.

## The Design Definitions Tree

One of the fundamental ideas in DSK was to use the filesystem as the interface for content creation. This enables _direct manipulation_ of the content and frees us from tedious administration interfaces.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk_fs.png)

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

## Building your own Frontend 

Architecture-wise DSK is split into a backend and frontend. The backend implemented 
in Go takes care of understanding the definitions tree and provides a REST API for the
frontend, usually implemented in JavaScript. 

Frontends are pluggable and the decoupled design allows you to create individually branded frontends. 
These are entirely free in their implementation, they must adhere to only a minimal set
of rules.

Read more about [building your own custom frontend](https://rundsk.com/tree/Architecture/Building-your-own-Frontend) and how to use with DSK.

## Contributing

Found a bug or just like to hack on DSK? We welcome contributions in many forms
to the DSK Open Source Project. It doesn't matter if you're a designer, Go
fairy, React sorceress or documentation fairy. 

Get in touch with us, by discussing your feature ideas, filing issues or fixing
bugs.

After following the _Setup_ procedure as outlined below a single command will
start DSK in development mode, ready for you to improve it.

```
$ make dev
```

The `make dev` command assumes your test design system definitions are
below the `example` directory. It also uses the built-in frontend from 
the `frontend/build` directory by default.

### Prerequisites

[Go](https://golang.org/) version 1.11 or later is needed for developing
and testing the application. [Node.js](https://nodejs.org) and
[Yarn](https://yarnpkg.com) are needed to build the built-in frontend.

If you're setting up a Go environment from scratch, add Go's `bin`
directory to your `PATH`, so that go binaries can be found. In
`.profile` add the following line.
```
export PATH=$PATH:$(go env GOPATH)/bin
```

DSK uses `go mod` to manage and vendor its dependencies. When using
Go 1.11 module support must be explictly enabled, add this line to
`.profile`:
```
export GO111MODULE=on
```

### Setup

Clone the official repository, switch into the directory and checkout the branch
you want to work on.

```
$ git clone github.com/atelierdisko/dsk 
$ git checkout 1.2
$ cd dsk
```

When using the built-in frontend, the `frontend/build` folder must be
present and contain the compiled version of the frontend. After a new
checkout you can create it, using the following command. 

```
$ make -C frontend build
```

### Improving the frontend

There are two ways to work on the built-in frontend and verify changes
in your browser easily at the same time while you go.

First start the backend which will also serve the frontend.
```
$ make dev
```

Each time you change the source of the frontend run the following
command and reload the browser window to see the changes.

```
$ make -C frontend build
```

Alternatively you can start the frontend (using a development server)
and the backend separately on different ports, having the frontend use
backend over its HTTP API.

Start the backend, first. By default it'll be reachable over port 8080.
```
$ make dev
```

Second, start the frontend using a development server. It will be
reachable over port 3000 and proxy request through to the backend.

```
$ cd frontend
$ yarn start
```

Now open http://127.0.0.1:3000 in your browser.

### Testing

To run the unit tests for the backend use `make test`, to run the
benchmarks use `make bench`, for performance profiling run `make profile`.

### Distributing

When updating the files of the built-in frontend, the file `frontend_vfsdata.go`
also needs to be remade. Run `make frontend_vfsdata.go` to do so. The file is 
provided to make DSK go gettable.

```
$ make -C frontend build
$ make frontend_vfsdata.go
$ VERSION=1.2.0-beta make dist
```

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.
