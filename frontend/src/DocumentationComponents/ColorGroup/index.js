/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import './ColorGroup.css';
import { Client } from '@rundsk/js-sdk';
import ColorCard from '../ColorCard';

function ColorGroup(props) {
  const [colors, setColors] = useState([]);

  useEffect(() => {
    if (props.src) {
      Client.fetch(props.src).then((data) => setColors(data.colors));
    }
  }, [props.src]);

  let classes = ['color-group'];

  let content = props.children;

  if (props.compact) {
    classes.push('color-group--is-compact');

    // We have to make sure the compact property is set on all
    // children as well
    content = React.Children.map(props.children, (c) => {
      return React.cloneElement(c, { compact: true });
    });
  }

  // If the src prop is set, the information about the colors should
  // be loaded via the API
  if (props.src) {
    content = colors.map((c) => {
      return (
        <ColorCard key={c.value} color={c.value} comment={c.comment} id={c.id} compact={props.compact}>
          {c.name}
        </ColorCard>
      );
    });
  }

  return <div className={classes.join(' ')}>{content}</div>;
}

export default ColorGroup;
