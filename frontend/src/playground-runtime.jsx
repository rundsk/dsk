/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useLayoutEffect } from 'react';
import ReactDOM from 'react-dom';

const PlaygroundWrapper = () => {
  useLayoutEffect(() => {
    const id = document.querySelector('body').getAttribute('data-id');
    window.parent.postMessage(
      JSON.stringify({
        id,
        contentHeight: document.querySelector('html').offsetHeight,
      }),
      '*'
    );
  });

  return (
    <div
      style={{
        width: '100%',
        paddingTop: 48,
        paddingBottom: 48,
        paddingLeft: 64,
        paddingRight: 64,
      }}
    >
      <ThePlaygroundInQuestion />
    </div>
  );
};

document.addEventListener('DOMContentLoaded', () => {
  ReactDOM.render(<PlaygroundWrapper />, document.getElementById('root'));
});
