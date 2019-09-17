/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import ReactDOM from 'react-dom';
import PAge from '.';
import { mount } from 'enzyme';
import 'jest-enzyme';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<Page />, div);
});


