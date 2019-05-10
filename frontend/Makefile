# Copyright 2019 Atelier Disko. All rights reserved.
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

build: node_modules $(shell find src)
	yarn run build

node_modules: yarn.lock
	yarn install
