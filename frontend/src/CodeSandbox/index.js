/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import './CodeSandbox.css';

function CodeSandbox(props) {
  return (
    <div className="codesandbox">
      <iframe
        title={props.id}
        allow="geolocation; microphone; camera; midi; vr; accelerometer; gyroscope; payment; ambient-light-sensor; encrypted-media; usb"
        sandbox="allow-modals allow-forms allow-popups allow-scripts allow-same-origin"
        src={`https://codesandbox.io/embed/${props.id}?fontsize=12&view=${props.view}&module=${props.file}&view=${props.view}`}
      />
    </div>
  );
}

export default CodeSandbox;
