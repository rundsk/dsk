/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import { BaseLink, withRoute } from 'react-router5';
import { Client } from '@atelierdisko/dsk';
import './NodeList.css';
import Tags from '../Tags';
import Heading from '../Heading';

function Node(props) {
  const [data, setData] = useState(null);

  useEffect(() => {
    getData();
  }, [props.url]);

  function getData() {
    if (props.url) {
      Client.get(props.url).then(data => {
        setData(data);
      });
    }
  }

  return (
    <BaseLink router={props.router} routeName="node" routeParams={{ node: `${props.url}` }} className="node-list__node">
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
          return <Node {...n} key={n.url} router={props.router} />;
        })}
    </>
  );
}

export default withRoute(NodeList);
