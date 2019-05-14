# Docker

We make official prebuilt docker container images 
[available on docker hub](https://cloud.docker.com/u/atelierdisko/repository/registry-1.docker.io/atelierdisko/dsk).

You can run a DSK container, by mounting the _Design Definitions Tree_, that is
the directory containing your design system documents, into the container. By
using the `DDT_LANG` enviroment variable, you specify which language the
search indexing should use, when looking at your design documents. 

```
docker run --rm -it \
	--expose :80:8080 \
	--mount type=bind,source="/path/to/ddt",target=/ddt \
	--env DDT_LANG=en \
	atelierdisko/dsk:1.2.0-beta
```

## With a Custom Frontend

The build expects to access the frontend under `./frontend`,
from where it copies, builds and then compiles it into the binary.

The `Dockerfile` hardcodes the process to build the built-in frontend. However
it can be reused for your custom frontend without any changes, when your
frontend...

1. ...is stored next to the `Dockerfile` in a `frontend` directory, 
2. ...has a `yarn.lock`, so `yarn install` installs all dependencies 
3. ...will create a production build, when `yarn run build` is executed, 
4. ...will store the build artifacts in a `build` subfolder.

In this case you remove the built-in frontend folder and replace it with yours:

```
rm -r frontend
git clone git@github.com:acme/frontend frontend
```

Then build a new container image, tagging it appropriately. The `VERSION` build
argument is supported, which sets the version compiled into DSK.

```
docker build \
	--tag acme/dsk:1.2.3 \
	--build-arg VERSION=1.2.3+acme \
	.
```

If your frontend requires a different build process, we recommend to modifiy the
`Dockerfile` and change it so it suits your needs.

