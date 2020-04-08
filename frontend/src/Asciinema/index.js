/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import './Asciinema.css';

function Asciinema(props) {
  return (
    <div className="asciinema">
      <iframe src={`https://asciinema.org/a/${props.id}/iframe`} />
    </div>
  );
}

export default Asciinema;
