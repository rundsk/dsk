/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';

import './Image.css';

// Images whose `src` attribute includes `@2x` will get their height and width
// set via CSS, so they are displayed half their size. Information about the
// natural dimensions is set by the DSK backend.
function Image(props) {
  let width = props.width;
  let height = props.height;

  if ((props.src.includes('@2x') || props.src.includes('@x2')) && width && height) {
    width /= 2;
    height /= 2;
  }
  return (
    <figure className="image">
      <img alt={props.alt} src={props.src} width={width} height={height} />

      {props.caption && <figcaption className="image__caption">{props.caption}</figcaption>}
    </figure>
  );
}

export default Image;
