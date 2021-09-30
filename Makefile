# Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
# Copyright 2017 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FRONTEND ?= $(shell pwd)/frontend/build
DDT ?= $(shell pwd)/../example-design-system

VERSION ?= head-$(shell git rev-parse --short HEAD)
LDFLAGS = -X main.Version=$(VERSION)
ANY_DEPS = $(shell find cmd internal)
ALL_PKGS = $(shell go list ./...)
CMD_PKG = github.com/rundsk/dsk/cmd/dsk

.PHONY: test
test:
	go test -tags=dev -race -ldflags "$(LDFLAGS)" $(ALL_PKGS)

.PHONY: bench
bench:
	go test -tags=dev -run XXX -bench -ldflags "$(LDFLAGS)" .

.PHONY: profile
profile:
	go test -tags=dev -run ^$$ -bench . -cpuprofile cpu.prof -memprofile mem.prof -mutexprofile mutex.prof $(ALL_PKGS)

.PHONY: lint
lint: internal/frontend/vfsdata.go
	go vet -all $(shell go list ./...)

.PHONY: dev
dev: internal/frontend/vfsdata.go
	go build -tags=dev -ldflags "$(LDFLAGS)" $(CMD_PKG)
	./dsk -frontend $(FRONTEND) "$(DDT)"
	rm dsk

.PHONY: clean
clean:
	if [ -f ./internal/frontend/vfsdata.go ]; then rm ./internal/frontend/vfsdata.go; fi
	if [ -d ./dist ]; then rm -r ./dist; fi
	if [ -f ./dsk ]; then rm ./dsk; fi
	if [ -f ./dsk.test ]; then rm ./dsk.test; fi
	if [ -f ./cpu.prof ]; then rm ./cpu.prof; fi
	if [ -f ./mem.prof ]; then rm ./mem.prof; fi

.PHONY: dist
dist: dist/dsk-darwin-amd64 dist/dsk-linux-amd64 dist/dsk-windows-386.exe 
dist: dist/dsk-darwin-amd64.zip dist/dsk-linux-amd64.tar.gz dist/dsk-windows-386.zip
dist: container-image
	ls -lh dist

.PHONY: container-image
container-image:
	docker buildx build --tag rundsk/dsk:$(VERSION) --build-arg VERSION=$(VERSION) .

.PHONY: container-push
container-push: container-image
	docker push rundsk/dsk:$(VERSION)

dist/%.zip: dist/%
	cd dist && zip $(notdir $@) $(notdir $<)
	cd example && zip -r ../$@ .

dist/dsk-windows-386.zip: dist/dsk-windows-386.exe
	cd dist && zip $(notdir $@) $(notdir $<)
	cd example && zip -r ../$@ .

dist/%.tar: dist/%
	cd dist && tar -cpvf $(notdir $@) $(notdir $<)
	cd example && tar -rvf ../$@ .

dist/%.tar.gz: | dist/%.tar
	gzip $(basename $@)

dist/%-darwin-amd64: $(ANY_DEPS) internal/frontend/vfsdata.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o $@ $(CMD_PKG)

dist/%-linux-amd64: $(ANY_DEPS) internal/frontend/vfsdata.go
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o $@ $(CMD_PKG)

dist/%-windows-386.exe: $(ANY_DEPS) internal/frontend/vfsdata.go
	GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS) -s -w" -o $@ $(CMD_PKG)

internal/frontend/vfsdata.go: $(shell find $(FRONTEND) -type f) 
	FRONTEND=$(FRONTEND) go run cmd/frontend/generate.go
