/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import './Meta.css';

function Meta(props) {
  return (
    <div className="meta">
      <div className="meta__key">{props.title}</div>
      <div className="meta__value">{props.children}</div>
    </div>
  );
}

export default Meta;
