/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import Link from '../Link';
import './Breadcrumbs.css';

function Breadcrumbs(props) {
  let crumbs;

  if (props.crumbs) {
    crumbs = props.crumbs.map((c) => {
      return (
        <li className="breadcrumbs__crumb" key={c.title}>
          <Link to={`/${c.url}`} key={'link'}>
            {c.title}
          </Link>
        </li>
      );
    });

    crumbs.pop();
  }
  return <ul className="breadcrumbs">{crumbs}</ul>;
}

export default Breadcrumbs;
