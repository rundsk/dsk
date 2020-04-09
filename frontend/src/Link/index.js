/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import { BaseLink, withRoute } from 'react-router5';

// Replace links to internal node with links from the router.
function Link(props) {
  let node = props['data-node'];
  let nodeAsset = props['data-node-asset'];

  // Do not use router for non nodes (i.e. external links)
  // or assets (these should be downloaded directly).
  if (!node || nodeAsset) {
    return (
      <a href={props.href} target={props.target}>
        {props.children}
      </a>
    );
  }
  // Create an URL Instance from the href String and create an URLSearchParams Instance from it.
  // Cast the URLSearchParams to an Object to merge later on. This way we can keep arbitary query parameters.
  let url = new URL(props.href, window.location.origin);
  let urlParams = Object.fromEntries(new URLSearchParams(url.search));

  return (
    <BaseLink
      router={props.router}
      routeName="node"
      routeParams={{ ...urlParams, node, v: props.route.params.v }}
      target={props.target}
    >
      {props.children}
    </BaseLink>
  );
}

export default withRoute(Link);
