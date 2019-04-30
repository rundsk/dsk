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

Read more about [the Design Definitions Tree](https://rundsk.com/tree/The-Design-Definitions-Tree) on our website, and how to add meta data, assets and authors.

## Building your own Frontend 

Architecture-wise DSK is split into a backend and frontend. The backend implemented 
in Go takes care of understanding the definitions tree and provides a REST API for the
frontend, usually implemented in JavaScript. 

Frontends are pluggable and the decoupled design allows you to create individually branded frontends. 
These are entirely free in their implementation, they must adhere to only a minimal set
of rules.

Read more about [building your custom frontend](https://rundsk.com/tree/Architecture/Building-your-own-Frontend) on our website and how to use together with DSK.

## Development

[Go](https://golang.org/) version 1.11 or later is needed for developing and testing the application.

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

The `make dev` command assumes your test design system definitions are
below the `example` directory.

```
$ go get github.com/atelierdisko/dsk
$ cd $(go env GOPATH)/src/github.com/atelierdisko/dsk
$ make dev
```

To run the unit tests use `make test`, to run the benchmarks use `make bench`,
for performance profiling run `make profile`.

When updating the files of the built-in frontend, the file `frontend_vfsdata.go`
also needs to be remade. Run `make frontend_vfsdata.go` to do so. The file is 
provided to make DSK go gettable.

## Known Limitations

- Symlinks inside the design definitions tree are not supported
- Not tested thoroughly on Windows

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.
