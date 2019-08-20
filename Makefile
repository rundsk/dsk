# Copyright 2017 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FRONTEND ?= $(shell pwd)/frontend/build
DDT ?= $(shell pwd)/example
DDT_LANG ?= en

VERSION ?= head-$(shell git rev-parse --short HEAD)
LDFLAGS = -X main.CompiledVersion=$(VERSION)
ANY_DEPS = $(wildcard *.go)

.PHONY: test
test:
	go test -tags=dev -race -ldflags "$(LDFLAGS)"

.PHONY: bench
bench:
	go test -tags=dev -run XXX -bench -ldflags "$(LDFLAGS)" .

.PHONY: profile
profile:
	go test -tags=dev -run ^$$ -bench -ldflags "$(LDFLAGS)" . -cpuprofile cpu.prof -memprofile mem.prof
	@echo Now run: go tool pprof dsk.test cpu.prof

.PHONY: dev
dev:
	go build -mod=vendor -tags=dev -race -ldflags "$(LDFLAGS)"
	./dsk -lang $(DDT_LANG) -frontend $(FRONTEND) $(DDT)

.PHONY: clean
clean:
	if [ -d ./dist ]; then rm -r ./dist; fi
	if [ -f ./dsk ]; then rm ./dsk; fi
	if [ -f ./dsk.test ]; then rm -r ./dsk.test; fi
	if [ -f ./cpu.prof ]; then rm -r ./cpu.prof; fi
	if [ -f ./mem.prof ]; then rm -r ./mem.prof; fi

.PHONY: dist
dist: dist/dsk-darwin-amd64 dist/dsk-linux-amd64 dist/dsk-windows-386.exe 
dist: dist/dsk-darwin-amd64.zip dist/dsk-linux-amd64.tar.gz dist/dsk-windows-386.zip
dist: container-image
	ls -lh dist

.PHONY: container-image
container-image:
	docker build --tag atelierdisko/dsk:$(VERSION) --build-arg VERSION=$(VERSION) .

.PHONY: container-push
container-push: container-image
	docker push atelierdisko/dsk:$(VERSION)

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

dist/%-darwin-amd64: $(ANY_DEPS) frontend_vfsdata.go
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS) -s -w" -o $@

dist/%-linux-amd64: $(ANY_DEPS) frontend_vfsdata.go
	GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS) -s -w" -o $@

dist/%-windows-386.exe: $(ANY_DEPS) frontend_vfsdata.go
	GOOS=windows GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS) -s -w" -o $@

frontend_vfsdata.go: $(shell find $(FRONTEND) -type f) 
	FRONTEND=$(FRONTEND) go run -mod=vendor frontend_generate.go
