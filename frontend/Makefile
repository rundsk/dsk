# Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
# Copyright 2019 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY: dev
dev:
	yarn react-scripts start

.PHONY: test
test:
	yarn test

.PHONY: build
build:
	yarn react-scripts build .

.PHONY: dist
dist:

.PHONY: lint
lint:
	yarn eslint --cache src

.PHONY: format
format:
	yarn eslint --fix --cache src
	yarn prettier --write 'src/**/*.{js,jsx,css,scss,json}'

.PHONY: clean
clean:
	if [ -d ./build ]; then rm -r ./build; fi
	if [ -d ./dist ]; then rm -r ./dist; fi
