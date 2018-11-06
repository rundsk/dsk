# Builtin DSK Frontend

This is the frontend being used by default by DSK. 

We are using vanilla JavaScript utilizing modern web technologies to build
it, so that - first - the code can be understood without expert knowledge of
specific frameworks, libraries or (sometimes very complicated) transpilation
processes. We hope that custom frontend authors find the code a good kind of
documentation when building their own. Second - we want to test how far this
development concept can take us today.

The frontend interacts with the DSK backend using its HTTP API. It accesses
that API via the `Client` JavaScript class from the DSK JavaScript package in
`js/dsk`. 

The DSK JavaScript package contains other usefull utilities, like `Tree`, that
help with traversing and filtering trees, i.e. for the filter navigation. The
package can be separately installed via `npm install @atelierdisko/dsk`.

## Requirements

DSK APIv2 and a browser that supports ES6 and ES modules.

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise stated. Use of the
source code is governed by a BSD-style license that can be found in the LICENSE
file.


