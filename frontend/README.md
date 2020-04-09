# Built-in DSK Frontend

This is the frontend being used by default by DSK. We hope that custom frontend
authors find the code a good kind of documentation and source of inspiration
when building their own.

The frontend was bootstrapped with [Create React
App](https://github.com/facebook/create-react-app).

The frontend interacts with the DSK backend using its HTTP API. It accesses
that API via the `Client` JavaScript class from the [DSK JavaScript
package](https://www.npmjs.com/package/@rundsk/js-sdk).

## Requirements

DSK APIv2, [Node.js](https://nodejs.org) and [Yarn](https://yarnpkg.com) are
needed to build the built-in frontend.

## Development

Dependencies must first be installed via `yarn install`.

Then `make dev`, runs the app in the development mode.
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

`make test`, launches the test runner in the interactive watch mode. See the 
section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) 
for more information.

`make dist`, builds the app for production to the `build` folder. It
correctly bundles React in production mode and optimizes the build for the best
performance. The build is minified and the filenames include the hashes.
Your app is ready to be deployed with a DSK build!

To run the frontend source code through `prettier`, use `make prettier`.
