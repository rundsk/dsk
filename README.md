# Design System Kit (DSK)

## Abstract

Using Design System Kit (DSK) you quickly organize components into a
browsable and live-searchable component library.

Hierachies between components are established using plain simple directories. Creating
documentation is as easy as adding a Markdown formatted file to a directory inside the _design definitions tree_.

![screenshot](https://atelierdisko.de/assets/app/img/github_dsk.png)

## Status

DSK is currently in development and **should not yet be considered for general production use**. 
Please see the _Development_ section for how to build the tool from source. We encourage contributions in form of code, design ideas or testing.

## Quickstart

[Download the binary](https://github.com/atelierdisko/hoi/releases) from GitHub and make it executable.

```
curl -L https://github.com/atelierdisko/dsk/releases/download/v0.5.0/dsk-darwin-amd64 -o dsk
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
│   │   ├── main.css
│   │   ├── main.js
│   │   ├── readme.md
│   │   └── unmask.svg
```

## Component Configuration File

Each component directory may hold an `index.json` file. The file and any of its
configuration settings are entirely optional.

Using the configuration file, we can add meta data to the component (i.e. keywords)
to improve the search experience in the interface. We can also configure
so called `demos` that describe variants of the component.

An example of a full component configuration looks like this:

```json
{
    "description": "This is a very very very fancy component.",
    "keywords": ["fancy", "very"],
    "import": "InputPassword",
    "demos": {
        "23 is the magic number": {"bar": "baz", "23": true},
        "Another Example": {"bla": "44"},
        "What happens when qux equals sup": {"qux": "sup"}
    }
}
```

Possible configuration options are:

- `description`: A single sentence that roughly describes the component.
- `keywords`: An array of keywords to group related components together.
- `import`: The name under which the component can be imported. By default the
   import name is the same as the directory path, i.e. `DataEntry/TextField/Password`. When your
   component is known under a different name, this option allows you to override
   the default one.
- `demos`: An object, which defines variations of the component. The object's
  keys are the names of the variations, the corresponding values hold property sets.

## Use it with [Create React App](https://github.com/facebookincubator/create-react-app) 

After creating the react app, Modify the entry point in `src/index.js`. When
loaded in DSK it must provide a `renderComponent()` function. This global
function is the only little glue code required. 

```javascript
window.renderComponent = function(root, name, props) {
  return import('./' + name).then(c => {
    ReactDOM.render(React.createElement(c.default, props), root);
  });
};
```

Using the `INSIDE_DSK` constant, it's possible to conditionally activate certain
default behavior while developing and deactivate it when embedded in DSK.

```javascript
if (typeof INSIDE_DSK === "undefined") {
  ReactDOM.render(<App />, document.getElementById('root'));
  registerServiceWorker();
}
```

When ready, create the bundle and copy it into our definitions tree.

```
$ npm run build 
$ cp build/static/*/main*.{css,js} example/DataEntry/TextField/Password/
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
directory called `_test`.

```
$ go get github.com/jteeuwen/go-bindata/...
$ go get github.com/atelierdisko/dsk
$ cd $(go env GOPATH)/src/github.com/atelierdisko/dsk
$ make dev
```

## Requirements

A filesystem and a browser that supports ES6.

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise
stated. Use of the source code is governed by a BSD-style
license that can be found in the LICENSE file.

