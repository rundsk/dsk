# Docker

We make official prebuilt docker container images
[available on docker hub](https://cloud.docker.com/u/atelierdisko/repository/registry-1.docker.io/atelierdisko/dsk).

The [Design Definitions Tree](/The-Design-Definitions-Tree), that is
the directory containing your design system documents, is not part of the container
image and must be mounted into the container when running it. The example below
uses a simple bind mount to make the documents available at the default `/ddt`
location inside the container.

By using the `DDT_LANG` environment variable, you specify which language the
search indexing should use, when looking at your design documents. In the example
below we use `en`.

The container exposes port `8080`, where it responds to incoming HTTP requests. In
our example we map it to port `80` on the docker host.

```
docker run --rm -it \
	--expose :80:8080 \
	--mount type=bind,source="/path/to/ddt",target=/ddt \
	--env DDT_LANG=en \
	atelierdisko/dsk:1.2.0
```
