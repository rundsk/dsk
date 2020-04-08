/**
 * Copyright 2020 Marius Wilms. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
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
