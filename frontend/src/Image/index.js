/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';

import './Image.css';

// Images whose `src` attribute includes `@2x` will get their height and width
// set via CSS, so they are displayed half their size. Information about the
// natural dimensions is set by the DSK backend.
function Image(props) {
  if (props.src.includes("@2x")) {
    if (props.width && props.height) {
      props.style = {
        maxWidth: `${props.width / 2}px`,
        maxHeight: `${props.height / 2}px`,
      };
    }
  }
  return <img className="image" alt={props.alt} src={props.src} style={props.style} />
}

export default Image;
