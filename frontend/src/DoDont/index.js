/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import './DoDont.css';
import ComponentDemo from '../ComponentDemo';

export function Do(props) {
  return (
    <>
      <div className="dodont-card__content">
        <div className="dodont-card__demo">
          <ComponentDemo background={props.background} backgroundcolor={props.backgroundcolor}>{props.children}</ComponentDemo>
        </div>
      </div>

      <div className="dodont-card__caption dodont-card__caption--do">
        <div className="dodont-card__title">Do</div>
        {props.caption}
      </div>
    </>
  )
}

export function Dont(props) {
  return (
    <>
      <div className="dodont-card__content">
        {props.strikethrough &&
          <svg>
            <line x1='-5%' y1='-5%' x2='105%' y2='105%' />
            <line x1='-5%' y1='105%' x2='105%' y2='-5%' />
          </svg>
        }
        <div className="dodont-card__demo">
          <ComponentDemo background={props.background} backgroundcolor={props.backgroundcolor}>{props.children}</ComponentDemo>
        </div>
      </div>

      <div className="dodont-card__caption dodont-card__caption--dont">
        <div className="dodont-card__title">Don’t</div>
        {props.caption}
      </div>
    </>
  )
}

function DoDont(props) {
  return (
    <div className="dodont">
      <div className="dodont__card-wrapper">
        {props.children}
      </div>
    </div>
  );
}

export default DoDont;
