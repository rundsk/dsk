# Copyright 2017 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

DEBUG ?= n

PREFIX ?= /usr/local
FLAG_PREFIX ?= $(PREFIX)

VERSION ?= head-$(shell git rev-parse --short HEAD)

GOFLAGS = -X main.Version=$(VERSION)

.PHONY: dev
dev:
	go-bindata -debug -o data.go data/...
	go build
	@if [ ! -d _test ]; then mkdir _test; fi
	./dsk _test

.PHONY: install
install: $(PREFIX)/bin/dsk

.PHONY: uninstall
uninstall:
	rm -f $(PREFIX)/bin/dsk

.PHONY: clean
clean:
	if [ -d ./dist ]; then rm -r ./dist; fi
	if [ -f ./dsk ]; then rm ./dsk; fi

.PHONY: dist
dist: dist/dsk dist/dsk-darwin-amd64 dist/dsk-linux-amd64

$(PREFIX)/bin/%: dist/%
	install -m 555 $< $@

dist/%: % data.go
	go build -ldflags "$(GOFLAGS)" -o $@ ./$<

dist/%-darwin-amd64: % data.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(GOFLAGS)" -o $@ ./$<

dist/%-linux-amd64: % data.go
	GOOS=linux GOARCH=amd64 go build -ldflags "$(_GOFLAGS)" -o $@ ./$<

data.go: $(shell find data -type f) 
ifeq ($(DEBUG),n)
	go-bindata -o data.go data/...
else
	go-bindata -debug -o data.go data/...
endif
