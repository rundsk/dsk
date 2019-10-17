# Copyright 2019 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY: dev
dev:
	yarn start

.PHONY: test
test:
	yarn test .

.PHONY: dist
dist: build

.PHONY: clean
clean:
	if [ -d ./node_modules ]; then rm -r ./node_modules; fi
	if [ -d ./build ]; then rm -r ./build; fi

build: node_modules $(shell find src)
	yarn run build

node_modules: yarn.lock
	yarn install

.PHONY: prettier
prettier: node_modules
	node_modules/.bin/prettier --write 'src/**/*.{css,js}'
