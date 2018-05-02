# Design System Kit

[![Build Status](https://travis-ci.org/atelierdisko/dsk.svg?branch=master)](https://travis-ci.org/atelierdisko/dsk)

## Abstract

Using the Design System Kit you quickly define and organize
_design aspects_ into a browsable and live-searchable design system.
Hierachies between design aspects are established using plain
simple directories. Creating documentation is as easy as adding a
[Markdown](https://guides.github.com/features/mastering-markdown/) formatted
file to a directory inside the _design definitions tree_.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk.png)

## Quickstart

1. Visit the [GitHub releases page](https://github.com/atelierdisko/dsk/releases) and download one of the quickstart packages for your operating system. For macOS use `dsk-darwin-amd64.zip`, for Linux use `dsk-linux-amd64.tar.gz`. 

2. The package is an archive that contains the `dsk` executable and an example design system. Double click on the downloaded file to unarchive both. 

3. On you can start dsk by double clicking on the excutable. On first use please follow [these instructions](https://support.apple.com/kb/PH25088) for macOS to skip the developer warning.

4. You should now see dsk starting in a small terminal window, [open the web application in your browser](http://localhost:8080), to browse through the design system.

_Alternatively_ the executable can be downloaded as a standalone binary. After downloading you must make the binary exectubale first, then execute it pointing to the directory containing the design definitions tree.

```
chmod +x dsk
./dsk example
```

## The Design Definitions Tree

One of the fundamental ideas in dsk was to use the filesystem as the interface for content creation. This enables _direct manipulation_ of the content and frees us from tedious adminstration interfaces.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk_fs.png)

The _design definitions tree_ (DDT for short), is a tree of
directories and subdirectories. Each of these directories stands for
a _design aspect_ in the hierarchy of your design system, these might
be actual components, when you are documenting the user interface, or
chapters of your company's guide into its design culture.

Each directory may hold several files to document these design aspects: a
configuration file, to add meta data or supporting _assets_ that can
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

_Note_: Directories beginning with an underscore (`_`), `x-` and `x_` or a dot (`.`) are ignored.

### Documenting Design Aspects

Aspects are documented by adding
[Markdown](https://guides.github.com/features/mastering-markdown/) formatted
documentation files to their directory. A `readme.md` file, may describe
an aspect or give clues how to use a certain component. You can split documentation
over several files when you like to. We usually use `api.md`, `explain.md` or
`comments.md`.

_Note_: `readme.md` is in no ways treated specially by dsk, but is usually displayed by
GitHub as the primary document in the web interface. 

_Another note_: If you prefer plain HTML documents over Markdown, these are
supported too. For this use `.html` instead of `.md` as the file
extension.

### Adding Aspect Meta Data (Description, Tags, ...)

To add more information to an aspect, we use a file called `meta.yml`. This 
file holds meta data, like a short description, tags and authors, about an aspect. 
The file uses the easy to write [YAML](https://www.youtube.com/watch?v=W3tQPk8DNbk) 
format.

_Note_: If you prefer to use [JSON](https://www.json.org) as a format,
that is supported too. Just exchange `.yml` for `.json` as the
extension.

An example of a full meta data file looks like this:

```yaml
authors: 
  - christoph@atelierdisko.de
  - marius@atelierdisko.de

description: > 
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.

keywords: 
  - typography
  - font
  - type

related:
  - DataEntry/TextField

tags:  
  - priority/1
  - release/0.1
  - progress/empty

version: 1.2.3
```

Possible meta data keys are:

- `authors`: An array of email addresses of the document authors; see below.
- `description`: A single sentence that roughly describes the design aspect.
- `keywords`: An array of terms that are searched in addition to `tags`.
- `related`: An array of related aspect URLs within dsk.
- `tags`: An array of tags to group related aspects together.
- `version`: A freeform version string.

### Authors

Each design aspect may be _authored_ by one or multiple humans. To assign
yourself, use the `authors` option in the `meta.yml` configuration file.

To enable automatic full names for each author, create an `AUTHORS.txt` file
inside the root of the DDT first. Each line of the file lists an author's full
name and her/his email address in angle brackets.

```text
Christoph Labacher <christoph@atelierdisko.de>
Marius Wilms <marius@atelierdisko.de>
```

### Manually Ordering Aspects and Documents

Aspects and documents by default appear in the same order as they are stored on your disk. But sometimes
order matters. To manually set the order you prefix aspects or documents with an _order number_ like so: 

```
example
├── DataEntry
│   ├── 01_TextField       <-- now comes before "Button"
│   │   ├── ...
│   │   ├── 01_explain.md  <-- now comes before "api.md"
│   │   └── 02_api.md
│   ├── 02_Button
│   │   └── ...
```

Valid order number prefixes look like `01_`, `01-`, `1_` or `1-`.

# Architecture

Architecture-wise dsk is split into a backend and frontend. The backend implemented 
in Go takes care of understanding the defintions tree and provides a REST API for the
frontend, usually implemented in JavaScript. 

Frontends are pluggable and the decoupled design allows you to create indvidually branded frontends. 
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

The backend provides the following API endpoints, that respond (with the
exception of assets) with JSON.

`GET /api/v1/hello`,
returns the version and a friendly greeting.

`GET /api/v1/tree`,
get the full design definitions tree as a nested tree of nodes.

`GET /api/v1/tree/{path}`,
get information about a single node specified by `path`.

`GET /api/v1/tree/{path}/{asset}`,
requests a node's asset, `{asset}` is a single filename.

`GET /api/v1/search?q={query}`,
performs a search over the design definitions tree and returns
a flat list of matched node URLs.

`GET /api/v1/messages`,
WebSocket for receiving messages from dsk, i.e. whenever the tree 
changes.

### Designing the URL Schema

Your frontend and its subdirectories will be mounted directly at the root path
`/`. Requests to anything under `/api` are routed to the backend, anything else
is routed into your application in `index.html`. 

Relative asset source paths inside documents will be made absolute to allow you
displaying a document's content wherever you like to.

### Baking

To _bake_ your frontend into dsk, install the the development tools as described in the _Development_ section frist. 
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

The `make dev` command assumes your test design system definitions are below
the `example` directory. The vendored dependencies are managed with
[dep](https://github.com/golang/dep).

```
$ go get -u github.com/golang/dep/cmd/dep
$ go get github.com/twitter/go-bindata/...
$ go get github.com/atelierdisko/dsk
$ cd $(go env GOPATH)/src/github.com/atelierdisko/dsk
$ make dev
```

To run the unit tests use `make test`, to run the benchmarks use `make bench`,
for performance profiling run `make profile`.

## Deploying as a Webservice

Once your design system is ready for the public, you surely want to share it
with co-workers inside your company or proudly present it to the whole world.

Design System Kit, works both in a local mode, as well in a hosted mode
on a webserver. There are two options for the hosted mode.

The following examples assume that you've installed the dsk binary
to `/bin/dsk`, are keeping the DDT in `/var/ds` and that
your operating system uses systemd as its init system.

### Simple

As long as you don't want any SSL encryption (you probably do), this is the
quickest way to get started. It keeps dsk running on your server and answers
requests directly.

Please replace `192.168.1.1` with the public IP address of
your machine. After [installing and starting the service
unit](https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units), the web interface should be available.


```ini
[Unit]
Description=Design System Kit

[Service]
ExecStart=/bin/dsk -host 192.168.1.1 -port 80 /var/ds
WorkingDirectory=/var/ds

[Install]
WantedBy=default.target
```

### With NGINX as a reverse-proxy and SSL

For SSL support we'll put dsk behind NGINX. The webserver will do the
termination for us, then forward all requests to dsk. Dsk will be listening on
the loopback interface on port 8080.

```ini
# ...

[Service]
ExecStart=/bin/dsk -port 8080 /var/ds
User=www-data
Group=www-data
WorkingDirectory=/var/ds

# ...
```

```nginx
server {
	listen 443 ssl http2;

	server_name example.com;
	root /var/ds;

	ssl_certificate /etc/ssl/certs/example.com.crt;
	ssl_certificate_key /etc/ssl/private/example.com.key;

	location / {
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $remote_addr;
		proxy_set_header Host $host;
		proxy_pass http://127.0.0.1:8080;
	}	
}
```

## Known Limitations

- Symlinks inside the design defintions tree are not supported
- Not tested thoroughly on Windows

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.
