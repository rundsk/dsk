/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import './Color.css';
import { copyTextToClipboard } from '../utils';
import contrast from 'get-contrast';

function ColorCard(props) {
  let classes = ['color'];

  let colorValue = props.color.toLowerCase();

  if (isColor(props.color) && contrast.ratio(colorValue, 'white') < 1.5) {
    classes.push('color--is-ultra-light');
  }

  function copyCode() {
    copyTextToClipboard(props.color);
  }

  return (
    <button className={classes.join(' ')} key={props.id} onClick={copyCode}>
      {props.children}
      <span className="color__demo" style={{ backgroundColor: props.color }}></span>
    </button>
  );
}

export default ColorCard;


const isColor = (strColor) => {
  const s = new Option().style;
  s.color = strColor;
  return s.color !== '';
}
