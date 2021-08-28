/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import { BaseLink, withRoute } from 'react-router5';
import './Breadcrumbs.css';

function Breadcrumbs(props) {
  let crumbs;

  if (props.crumbs) {
    crumbs = props.crumbs.map((c) => {
      return (
        <li className="breadcrumbs__crumb" key={c.title}>
          <BaseLink
            router={props.router}
            routeName="node"
            routeParams={{ node: `${c.url}`, v: props.route.params.v }}
            key={'link'}
          >
            {c.title}
          </BaseLink>
        </li>
      );
    });

    crumbs.pop();
  }
  return <ul className="breadcrumbs">{crumbs}</ul>;
}

export default withRoute(Breadcrumbs);
