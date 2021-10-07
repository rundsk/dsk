/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import './Banner.css';

function Banner(props) {
  let classes = ['banner'];

  switch (props.type) {
    case 'warning':
      classes.push('banner--warning');
      break;
    case 'error':
      classes.push('banner--error');
      break;
    case 'important':
      classes.push('banner--important');
      break;
    default:
      break;
  }

  return (
    <div className={classes.join(' ')}>
      {props.title && <div className="banner__header">{props.title}</div>}
      <div className="banner__content">{props.children}</div>
    </div>
  );
}

export default Banner;
