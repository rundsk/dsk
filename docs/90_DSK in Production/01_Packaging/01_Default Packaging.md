# Default Packaging

Usually all you need is, to download the right binary for your platform from
[the releases section of our project page](https://github.com/atelierdisko/dsk/releases).
The packaging documentation below gives you an in depth view into the DSK build process,
which matters most when creating a customized DSK build.

One of the main benefits of using Go is that we can
bundle everything except the DDT into one binary.

The build process is driven by GNU Make using the targets
defined in the `Makefile`. When executed directly on your system it requires to have some tools preinstalled.
Alternatively you can use [Docker to create a container image](Packaging?t=container-image).

Before continuing, we recommend making yourself familiar with the
[general architecture of DSK](/Architecture).

## Prerequisites

[Go](https://golang.org/) version 1.11 or later for the binary,
as well as [Node.js](https://nodejs.org) and [Yarn](https://yarnpkg.com)
are needed to build the built-in frontend.

## Supported environment variables

The build process can be customized through environment
variables:

- `VERSION` controls which version will be reported when later
running `dsk -version`. This can be any string, when not set it will default
to `head-` and the current short Git-ref, i.e. `head-123cafe`.

- `FRONTEND` controls which frontend will be _baked into_ DSK. By
default this is `frontend/build`, which will use the built-in frontend.

## Building the built-in frontend

Before the frontend can be compiled into the DSK binary in the next step, we
first must build the frontend itself.

```
make -C frontend dist
```

## Compiling the binaries

You compile a single binary for your platform with the following commands, this is faster
than building the whole world all the time.

```
make dist/dsk-linux-amd64  # ...for Linux
make dist/dsk-darwin-amd64 # ...for MacOS
```

Generally in preparation for a release, the following command will create
binaries for all supported platforms inside the `dist` folder, as well
as a docker container image.

```
make dist
VERSION=1.2.0 make dist # Preparing a release.
```

## Replacing the built-in frontend

To persistently _bake_ your frontend into DSK, effectively replacing the
built-in one, You create your custom DSK build by using the `FRONTEND`
environment variable.

First, please ensure your frontend is built and production ready assets along
with an `index.html` as an entrypoint have been created. In the example below we assume
these have been stored in a `build` subdirectory.

```
FRONTEND=/path/to/acme/frontend/build make dist
```

We recommend using a special version string to indicate that the DSK binary is
a custom one.

```
VERSION=1.2.0+acme FRONTEND=/path/to/acme/frontend/build make dist
```
