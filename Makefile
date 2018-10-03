# Copyright 2017 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

VERSION ?= head-$(shell git rev-parse --short HEAD)
GOFLAGS = -X main.Version=$(VERSION)
ANY_DEPS = $(wildcard *.go)
FRONTEND ?= $(shell pwd)/frontend

.PHONY: test
test:
	go test -tags=dev -race

.PHONY: bench
bench:
	go test -tags=dev -run XXX -bench .

.PHONY: profile
profile:
	go test -tags=dev -run ^$$ -bench . -cpuprofile cpu.prof -memprofile mem.prof
	@echo Now run: go tool pprof dsk.test cpu.prof

.PHONY: dev
dev:
	go build -tags=dev -race -ldflags "$(GOFLAGS)"
	./dsk example

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
	ls -lh dist

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
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(GOFLAGS) -s -w" -o $@

dist/%-linux-amd64: $(ANY_DEPS) frontend_vfsdata.go
	GOOS=linux GOARCH=amd64 go build -ldflags "$(GOFLAGS) -s -w" -o $@

dist/%-windows-386.exe: $(ANY_DEPS) frontend_vfsdata.go
	GOOS=windows GOARCH=386 go build -ldflags "$(GOFLAGS) -s -w" -o $@

frontend_vfsdata.go: $(shell find $(FRONTEND) -type f) 
	FRONTEND=$(FRONTEND) go run frontend_generate.go
