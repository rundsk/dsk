/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 *
 * @jest-environment jsdom
 */

import ReactDOM from 'react-dom';
import Page from '.';
import { GlobalContext } from '../App';

it('renders without crashing', () => {
  const div = document.createElement('div');
  let config, setConfig, source, setSource, filterTerm, setFilterTerm;

  ReactDOM.render(
    <GlobalContext.Provider
      value={{
        config,
        setConfig,
        source,
        setSource,
        filterTerm,
        setFilterTerm,
      }}
    >
      <Page />
    </GlobalContext.Provider>,
    div
  );
});
