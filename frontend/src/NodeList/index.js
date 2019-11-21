/**
 * Copyright 2020 Marius Wilms. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import { BaseLink, withRoute } from 'react-router5';
import { Client } from '@rundsk/dsk';
import './NodeList.css';
import Tags from '../Tags';
import Heading from '../Heading';

function Node(props) {
  const [data, setData] = useState(null);

  useEffect(() => {
    if (props.url) {
      Client.get(props.url, props.source).then(data => {
        setData(data);
      });
    }
  }, [props.url, props.source]);

  return (
    <BaseLink
      router={props.router}
      routeName="node"
      routeParams={{ node: `${props.url}`, v: props.route.params.v }}
      className="node-list__node"
    >
      <Heading className="node-list__node-title" level="beta" isJumptarget={false}>
        {props.title}
      </Heading>
      <div className="node-list__node-tags">
        <Tags tags={data && data.tags} isClickable={false} />
      </div>
      <div className="node-list__node-description">
        {data && data.description}
        <span className="node-list__node-children-count">
          {data && data.children.length > 0 && ` (${data.children.length} aspects)`}
        </span>
      </div>
      {/* <div className="node-list__cta">Details</div> */}
    </BaseLink>
  );
}

function NodeList(props) {
  return (
    <>
      {props.nodes &&
        props.nodes.map(n => {
          return <Node {...n} source={props.source} key={n.url} router={props.router} route={props.route} />;
        })}
    </>
  );
}

export default withRoute(NodeList);
