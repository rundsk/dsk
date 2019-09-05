# Building your own Frontend

DSK’s built-in frontend is written as a React app and is completely decoupled from the backend – it uses the open [API](../API) to communicate with it.

If you want a custom-styled frontend or need additional frontend features it is possible to build your own frontend using the same API and bundle it with the DSK binary. The default frontend is a good starting point when developing your own. It can be found in the `frontend` folder of this [project](http://github.com/atelierdisko/dsk).

## Available API Endpoints
Please read our [API document](../API) for in depth information.

## Designing the URL Schema
Your frontend and its subdirectories will be mounted directly at the root path `/`. Requests to anything under `/api` are routed to the backend, anything else is routed into your application in `index.html`.

Relative asset source paths inside documents will be made absolute to allow you displaying a document’s content wherever you like to.

## Starting DSK with the custom frontend
You “tell” DSK, to use your frontend by invoking it with the `frontend` flag, providing an absolute path or a path relative to the current working directory, that contains the frontend.

```shell
./dsk -frontend=/my/frontend example
```

## Replacing the built-in frontend
Please see our documentation on  [Packaging](/DSK-in-Production/Packaging) , to persistently _bake_ the custom frontend into the DSK binary. You either use the default way for doing this or use docker to create a container image.
