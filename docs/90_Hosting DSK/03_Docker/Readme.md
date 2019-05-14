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

You can also build your own images, if you like to. This works from the project
root and uses a multi stage build. 

Currently just the `VERSION` build argument is supported, which sets the version
compiled into DSK. The build expects to access the frontend under `./frontend`,
from where it copies, builds and then compiles it into the binary.

```
docker build \
	--tag acme/dsk:my-version \
	--build-arg VERSION=acme-my-version \
	.
```
