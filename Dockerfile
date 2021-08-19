# Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
# Copyright 2019 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Build stage to retrieve and build the frontend. Will pull in the frontend from
# the project root `frontend` folder. This stage hardcodes the build process for
# the built-in frontend. If you have a custom frontend with a different build
# process (although this one may work for you, as it's pretty standard), you'll
# need to create a new Dockerfile basing it off this one.
FROM alpine:latest as frontend

COPY frontend /src

RUN apk add nodejs
RUN apk add yarn

WORKDIR /src

RUN yarn
RUN yarn react-scripts build

# Build stage to compile the DSK binary, "baking in" the frontend from the
# previous stage. The frontend will replace the default frontend.
FROM golang:1.14 as backend

ARG VERSION=head
ENV GO111MODULE=on

COPY . /src
COPY --from=frontend /src/build /frontend

WORKDIR /src

# We cannot compile our go program with C support, otherwise results in a "not
# found" error, when running the binary. This also disables DDT change watching
# and possibly other (future) features. However it is not expected that the former
# feature is of any use in secenarios we imagine the docker
# container is being run. Alternatively the tutorial at
# http://kefblog.com/2017-07-04/Golang-ang-docker can be used to enable CGO
# support.
ENV CGO_ENABLED=0 

# Force make to re-generate the to-be-embedded assets file. After a fresh
# checkout of the source code all files have the same timestamp and the target
# will not be made, as it appears to be up to date. But in fact it contains
# entirely new files, especially when a custom frontend is used.
RUN touch /frontend/index.html

# Build through Makefile, as it i.e. ensures the frontend embed happens
# correctly and uses the right defaults for Go modules.
RUN FRONTEND=/frontend VERSION=$VERSION make dist/dsk-linux-amd64

# Final stage that executes the binary.
FROM alpine:latest as run

COPY --from=backend /src/dist/dsk-linux-amd64 /dsk

EXPOSE 8080

CMD /dsk -host 0.0.0.0 -port 8080 /ddt

