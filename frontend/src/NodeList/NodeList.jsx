/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import { Client } from '@rundsk/js-sdk';

import Tags from '../Tags';
import Heading from '../DocumentationComponents/Heading';
import Link from '../Link';

import './NodeList.css';

function Node(props) {
  const [data, setData] = useState(null);

  useEffect(() => {
    if (props.url) {
      Client.get(props.url, props.source).then((data) => {
        setData(data);
      });
    }
  }, [props.url, props.source]);

  return (
    <Link to={props.url} className="node-list__node">
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
    </Link>
  );
}

function NodeList(props) {
  return (
    <>
      {props.nodes &&
        props.nodes.map((n) => {
          return <Node {...n} source={props.source} key={n.url} />;
        })}
    </>
  );
}

export default NodeList;
