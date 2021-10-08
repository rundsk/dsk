/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import './DoDont.css';
import Playground from '../Playground';

export function Do(props) {
  return (
    <>
      <div
        className={`dodont-card__content dodont-card--do ${
          props.backgroundcolor === '#FFFFFF' ? 'dodont-card--white' : ''
        }`}
      >
        <div className="dodont-card__demo">
          <Playground background={props.background} backgroundcolor={props.backgroundcolor}>
            {props.children}
          </Playground>
        </div>
      </div>

      <div className="dodont-card__caption dodont-card__caption--do">
        <div className="dodont-card__title">Do</div>
        {props.caption}
      </div>
    </>
  );
}

export function Dont(props) {
  return (
    <>
      <div
        className={`dodont-card__content dodont-card--dont ${
          props.backgroundcolor === '#FFFFFF' ? 'dodont-card--white' : ''
        }`}
      >
        {props.strikethrough && (
          <svg>
            <line x1="-5%" y1="-5%" x2="105%" y2="105%" />
            <line x1="-5%" y1="105%" x2="105%" y2="-5%" />
          </svg>
        )}
        <div className="dodont-card__demo">
          <Playground background={props.background} backgroundcolor={props.backgroundcolor}>
            {props.children}
          </Playground>
        </div>
      </div>

      <div className="dodont-card__caption dodont-card__caption--dont">
        <div className="dodont-card__title">Don’t</div>
        {props.caption}
      </div>
    </>
  );
}

function DoDont(props) {
  return (
    <div className="dodont">
      <div className="dodont__card-wrapper">{props.children}</div>
    </div>
  );
}

export default DoDont;
