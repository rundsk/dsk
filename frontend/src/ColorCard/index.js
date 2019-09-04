/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState } from 'react';
import './ColorCard.css';
import { copyTextToClipboard } from '../utils';
import contrast from 'get-contrast';

function ColorCard(props) {
  const [showCopiedIndicator, setShowCopiedIndicator] = useState(false);

  let classes = ['color-card'];

  if (contrast.ratio(props.color, 'white') < 1.5) {
    classes.push('color-card--is-ultra-light');
  }

  if (props.compact) {
    classes.push('color-card--is-compact');
  }

  function copyCode() {
    setShowCopiedIndicator(true);
    copyTextToClipboard(props.color);

    setTimeout(() => {
      setShowCopiedIndicator(false);
    }, 1000);
  }

  return (
    <div className={classes.join(' ')} key={props.id} onClick={copyCode}>
      <div className="color-card__demo" style={{ backgroundColor: props.color }}>
        <div className="color-card__score">
          <span>{contrast.score(props.color, 'white')}</span>
          <span>{contrast.score(props.color, 'black')}</span>
        </div>

        <div
          className={`color-card__copied-indicator ${
            showCopiedIndicator ? 'color-card__copied-indicator--is-visible' : ''
          }`}
          style={{
            color: contrast.ratio(props.color, 'white') < contrast.ratio(props.color, 'black') ? 'black' : 'white',
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
    </div>
  );
}

export default ColorCard;
