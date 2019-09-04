/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

const proxy = require('http-proxy-middleware');

module.exports = function(app) {
  app.use(proxy('/api', {
    target: 'http://localhost:8080/',
    ws: true
  }));
};
