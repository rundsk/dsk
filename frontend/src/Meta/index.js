/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from "react";
import "./Meta.css";

function Meta(props) {
  return (
    <div className="meta">
      <div className="meta__key">{props.title}</div>
      <div className="meta__value">{props.children}</div>
    </div>
  );
}

export default Meta;
