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

## Setup

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

## Testing

To run the unit tests for the backend use `make test`, to run the
benchmarks use `make bench`, for performance profiling run `make profile`.

## Distributing

When updating the files of the built-in frontend, the file `frontend_vfsdata.go`
also needs to be remade. Run `make frontend_vfsdata.go` to do so. The file is 
provided to make DSK go gettable.

```
$ make -C frontend build
$ make frontend_vfsdata.go
$ VERSION=1.2.0-beta make dist
```
