/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useRef } from 'react';
import './ColorCard.css';
import { copyTextToClipboard } from '../utils';
import contrast from 'get-contrast';

function ColorCard(props) {
  const [showCopiedIndicator, setShowCopiedIndicator] = useState(false);
  const ref = useRef();

  let classes = ['color-card'];

  let colorValue = props.color.toLowerCase();

  if (isColor(props.color) && contrast.ratio(colorValue, 'white') < 1.5) {
    classes.push('color-card--is-ultra-light');
  }

  if (props.compact) {
    classes.push('color-card--is-compact');
  }

  function copyCode(ev) {
    ev.preventDefault();
    setShowCopiedIndicator(true);
    if (navigator.clipboard) {
      navigator.clipboard.writeText(props.color);
    } else {
      // navigator.clipboard does not work in Internet Explorer
      copyTextToClipboard(props.color);

      // Using copyTextToClipboard means the element loses focus. Here we restore it.
      if (ref.current) {
        ref.current.focus();
      }
    }

    setTimeout(() => {
      setShowCopiedIndicator(false);
    }, 1000);
  }

  return (
    <button className={classes.join(' ')} key={props.id} onClick={copyCode} ref={ref}>
      <div className="color-card__demo" style={{ backgroundColor: props.color }}>
        <div className="color-card__score">
          <span>{isColor(props.color) && contrast.score(colorValue, 'white')}</span>
          <span>{isColor(props.color) && contrast.score(colorValue, 'black')}</span>
        </div>

        <div
          className={`color-card__copied-indicator ${
            showCopiedIndicator ? 'color-card__copied-indicator--is-visible' : ''
          }`}
          style={{
            color:
              isColor(props.color) && contrast.ratio(colorValue, 'white') < contrast.ratio(colorValue, 'black')
                ? 'black'
                : 'white',
          }}
        >
          Copied!
        </div>
      </div>
      <div className="color-card__name">
        {props.children} <span className="color-card__id">({props.id})</span>
      </div>
      <div className="color-card__spec">{props.color}</div>
      <div className="color-card__comment">{props.comment}</div>
    </button>
  );
}

export default ColorCard;

const isColor = (strColor) => {
  const s = new Option().style;
  s.color = strColor;
  return s.color !== '';
};
