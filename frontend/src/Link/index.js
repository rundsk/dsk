/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import { BaseLink, withRoute } from 'react-router5';

// Replace links to internal node with links from the router.
function Link(props) {
  let node = props['data-node'];

  if (!node) {
    return <a href={props.href}>{props.children}</a>;
  }
  let hash = props.href.split('?t=')[1] || undefined;

  return (
    <BaseLink router={props.router} routeName="node" routeParams={{ node, t: hash }}>
      {props.children}
    </BaseLink>
  );
}

export default withRoute(Link);
