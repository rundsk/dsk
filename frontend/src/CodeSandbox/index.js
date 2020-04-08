/**
 *  Copyright 2019 Atelier Disko. All rights reserved. This source
 *  code is distributed under the terms of the BSD 3-Clause License.
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
