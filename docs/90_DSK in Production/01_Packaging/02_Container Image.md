# Container image

DSK comes with a multi stage `Dockerfile`. It allows you to perform DSK builds
without installing the build tools. The only prerequisite is that you have docker
installed and running.

A container image can be created using `docker build`.

```
docker build \
	--tag atelierdisko/dsk:1.2.3 \
	--build-arg VERSION=1.2.3 \
	. # <- Note the dot.
```

## Supported build `ARG`s

- `VERSION`, sets the version compiled into DSK.

## Using a custom frontend

The standard build expects to access the frontend under `./frontend`,
from wherefrom it copies, builds and then compiles it into the binary.

The `Dockerfile` hardcodes the process to build the built-in frontend. However
it can be reused for your custom frontend without any changes, when your
frontend...

1. ...is stored next to the `Dockerfile` in a `frontend` directory,
2. ...has a `yarn.lock`, so `yarn install` installs all dependencies
3. ...will create a production build, when `yarn run build` is executed,
4. ...will store the build artifacts in a `build` subfolder.

In this case you remove the built-in frontend folder and replace it with yours.

```
rm -r frontend
git clone git@github.com:acme/frontend frontend
```

Finally build a new container image, tagging it appropriately.

```
docker build \
	--tag acme/dsk:1.2.3 \
	--build-arg VERSION=1.2.3+acme \
	. # <- Note the dot.
```
