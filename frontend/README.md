# Built-in DSK Frontend

This is the frontend being used by default by DSK. We hope that custom frontend
authors find the code a good kind of documentation and source of inspiration
when building their own.

The frontend was bootstrapped with [Create React
App](https://github.com/facebook/create-react-app).

The frontend interacts with the DSK backend using its HTTP API. It accesses
that API via the `Client` JavaScript class from the [DSK JavaScript
package](https://www.npmjs.com/package/@atelierdisko/dsk). The package source
code can be found inside `js/dsk`.

The DSK JavaScript package contains other usefull utilities, like `Tree`, that
help with traversing and filtering trees, i.e. for the filter navigation. The
package can be separately installed via `npm install @atelierdisko/dsk`.

## Requirements

DSK APIv2, [Node.js](https://nodejs.org) and [Yarn](https://yarnpkg.com) are
needed to build the built-in frontend.

## Development

Dependencies must first be installed via `yarn install`.

Then `yarn start`, runs the app in the development mode.
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

`yarn test`, launches the test runner in the interactive watch mode. See the 
section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) 
for more information.

`yarn run build`, bilds the app for production to the `build` folder. It
correctly bundles React in production mode and optimizes the build for the best
performance. The build is minified and the filenames include the hashes.
Your app is ready to be deployed with a DSK build!

## Copyright & License

DSK is Copyright (c) 2017 Atelier Disko if not otherwise stated. Use of the
source code is governed by a BSD-style license that can be found in the LICENSE
file.


