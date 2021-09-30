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

First we'll be setting up an umbrella directory where the main dsk repository
and supporting repositories will live, which we'll clone into it.

```
mkdir rundsk
cd rundsk

git clone git@github.com:rundsk/dsk.git
git clone git@github.com:rundsk/js-sdk.git
git clone git@github.com:rundsk/example-component-library.git
git clone git@github.com:rundsk/example-design-system.git
```

All the following documentation will assume that you are working from inside the
main repository checkout.

```
cd dsk
```

## Developing

First we'll start the backend in development mode and after that do the same
with the frontend.

On a fresh checkout we'll have to install the dependencie of the frontend
first, and then start the frontend's development server.

```
cd frontend
yarn
make dev
```

You should now be able to reach the frontend when opening http://127.0.0.1:3000
in your browser. You might see some errors as we not yet have started the backend.

Now inside another terminal we'll start the backend.

```
make dev
```

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
