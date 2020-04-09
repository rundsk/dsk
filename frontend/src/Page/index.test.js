/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import ReactDOM from 'react-dom';
import Page from '.';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<Page />, div);
});
