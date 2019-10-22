/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import ReactDOM from 'react-dom';
import Page from '.';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<Page />, div);
});
