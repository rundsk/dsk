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
  // Create an URL Instance from the href String and create an URLSearchParams Instance from it.
  // Cast the URLSearchParams to an Object to merge later on. This way we can keep arbitary query parameters.
  let url = new URL(props.href, window.location.origin);
  let urlParams = Object.fromEntries(new URLSearchParams(url.search));

  return (
    <BaseLink router={props.router} routeName="node" routeParams={{ ...urlParams, node, v: props.route.params.v }}>
      {props.children}
    </BaseLink>
  );
}

export default withRoute(Link);
