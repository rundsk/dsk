/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import './Glitch.css';

function Glitch(props) {
  return (
    <div className="glitch">
      <iframe title={`glitch-embed-${props.id}`} src={`https://glitch.com/embed/#!/embed/${props.id}?path=${props.file}&previewSize=0`} />
    </div>
  );
}

export default Glitch;
