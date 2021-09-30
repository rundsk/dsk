# Contributing

Found a bug or just like to hack on DSK? We welcome contributions in many forms
to the DSK Open Source Project. It doesn't matter if you're a designer, Go
fairy, React sorceress or documentation wizard. 

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

## Prerequisites

[Go](https://golang.org/) version 1.17 or later is needed for developing
and testing the application. [Node.js](https://nodejs.org) and
[Yarn](https://yarnpkg.com) are needed to build the built-in frontend.

If you're setting up a Go environment from scratch, add Go's `bin`
directory to your `PATH`, so that go binaries can be found. In
`.profile` add the following line.
```
export PATH=$PATH:$(go env GOPATH)/bin
```

## Setup

Clone the official repository, switch into the directory and checkout the branch
you want to work on.

```
$ git clone github.com/rundsk/dsk 
$ git checkout 1.2
$ cd dsk
```

When using the built-in frontend, the `frontend/build` folder must be
present and contain the compiled version of the frontend. After a new
checkout you can create it, using the following command. 

```
$ make -C frontend build
```

## Improving the frontend

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
reachable over port 3000 and proxy requests through to the backend.

```
$ make -C frontend dev
```

Now open http://127.0.0.1:3000 in your browser.

## Debugging

When running the tests the search indexes are store to disk, to ease
debugging search issues. The test will output the paths where the
indexes are stored.

The `bleve` tool can then be used to interact with the index files directly.
The tool can be installed with `go get github.com/blevesearch/bleve/...`. The
`bleve` command should than be available.

## Testing

To run the unit tests for the backend use `make test`, to run the
benchmarks use `make bench`, for performance profiling run `make profile`.

## Distributing

Before distrubtiong the binary, the data file which contains the built-in
frontend assets must be made. This happens automatically as it is a dependency
of the dist target.

```
$ make -C frontend build
$ VERSION=1.2.0-beta make dist
```
