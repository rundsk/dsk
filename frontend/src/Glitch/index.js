/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import './Glitch.css';

function Glitch(props) {
  return (
    <div className="glitch">
      <iframe src={`https://glitch.com/embed/#!/embed/${props.id}?path=${props.file}&previewSize=0`} />
    </div>
  );
}

export default Glitch;
