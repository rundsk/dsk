/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import 'regenerator-runtime/runtime';
import Enzyme from 'enzyme';
import Adapter from '@wojtekmaj/enzyme-adapter-react-17';
import React from 'react';

global.React = React;
Enzyme.configure({ adapter: new Adapter() });
