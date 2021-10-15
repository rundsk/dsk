# Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
# Copyright 2019 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

NODE_ENV ?= production
MINIFY ?= y

ALL_SOURCE = $(shell find src -type d \( -name build -or -name dist \) -prune -or -type f -print)
ALL_PUBLIC = $(addprefix build/,$(notdir $(wildcard public/*)))

.PHONY: dev
dev: $(ALL_PUBLIC)
	SERVE=y MINIFY=n NODE_ENV=development yarn node esbuild.js

.PHONY: test
test:
	NODE_ENV=test yarn jest

build/static: $(ALL_SOURCE)
	@mkdir -p build/static
	MINIFY=$(MINIFY) NODE_ENV=$(NODE_ENV) yarn node esbuild.js

build/%: public/%
	@mkdir -p build
	cp $< $@

.PHONY: build
build: build/static $(ALL_PUBLIC) 
	
.PHONY: dist
dist:

.PHONY: lint
lint:
	yarn eslint --cache 'src/**/*.js{,x}'

.PHONY: format
format:
	yarn eslint --fix --cache 'src/**/*.js{,x}'
	yarn prettier --write 'src/**/*.{js,jsx,css,scss,json}'

.PHONY: clean
clean:
	if [ -d ./build ]; then rm -r ./build; fi
	if [ -d ./dist ]; then rm -r ./dist; fi
