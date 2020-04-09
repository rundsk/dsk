/**
 * Copyright 2020 Marius Wilms. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import ReactDOM from 'react-dom';

import * as serviceWorker from './serviceWorker';
import createRouter from 'router5';
import browserPlugin from 'router5-plugin-browser';
import { RouterProvider } from 'react-router5';
import { setGlobal } from 'reactn';

import './index.css';
import App from './App';

const routes = [
  { name: 'home', path: '/?:v' },
  { name: 'node', path: '/tree/*node?:t&:v' },
];

const router = createRouter(routes, {
  defaultRoute: 'home',
});

router.usePlugin(
  browserPlugin({
    useHash: false,
  })
);

router.start();

setGlobal({
  filterTerm: '',
  source: '',
  config: {
    _populated: false,
  },
});

ReactDOM.render(
  <RouterProvider router={router}>
    <App />
  </RouterProvider>,
  document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
