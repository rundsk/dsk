/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

const { pnpPlugin } = require('@yarnpkg/esbuild-plugin-pnp');
const esbuild = require('esbuild'); // eslint-disable-line import/no-self-import
const http = require('http');

let bundle = {
  entryPoints: ['src/index.jsx'],
  bundle: true,
  splitting: true,
  sourcemap: true,
  format: 'esm',
  target: ['es2020'],
  minify: process.env.MINIFY === 'y',
  plugins: [pnpPlugin()],
  outdir: 'build/static',
  loader: { '.svg': 'dataurl', '.png': 'file' },
  assetNames: 'assets/[name]-[hash]',
  chunkNames: 'chunks/[name]-[hash]',
  platform: 'browser',
  define: {
    'process.env.NODE_ENV': `'${process.env.NODE_ENV}'`,
  },
};
if (process.env.WATCH === 'y') {
  esbuild.build({ ...bundle, color: true, logLevel: 'debug', watch: true }).catch(() => process.exit(1));
} else if (process.env.SERVE === 'y') {
  esbuild
    .serve(
      {
        servedir: 'build',
      },
      bundle
    )
    .then((frontend) => {
      http
        .createServer((oreq, ores) => {
          oreq.pipe(
            http.request(
              {
                hostname: 'localhost',
                port: oreq.url.startsWith('/api/') ? 8080 : frontend.port,
                path: oreq.url,
                method: oreq.method,
                headers: oreq.headers,
              },
              (pres) => {
                ores.writeHead(pres.statusCode, pres.headers);
                pres.pipe(ores, { end: true });
              }
            ),
            { end: true }
          );
        })
        .listen(3000);

      console.log('Started frontend development server, please visit: http://localhost:3000');
      console.log('(Expecting backend at localhost:8000 for API request routing.)');
    });
} else {
  esbuild.build({ ...bundle, color: true, logLevel: 'debug' }).catch(() => process.exit(1));
}
