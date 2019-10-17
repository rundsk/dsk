/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import './SourceView.css';
import { Client } from '@atelierdisko/dsk';
import CodeBlock from '../CodeBlock';

function SourceView(props) {
  const [data, setData] = useState(null);

  useEffect(() => {
    if (props.url === 'hello') {
      Client.hello().then(data => {
        setData(data);
      });
    } else if (props.url !== undefined) {
      Client.get(props.url, props.source).then(data => {
        setData(data);
      });
    }
  }, [props.url, props.source]);

  let title = `API Response for /api/v2/tree/${props.url}`;
  console.log(props.source);

  if (props.source !== 'live') {
    title = `API Response for /api/v2/tree/${props.url} (Source: ${props.source})`;
  }

  if (props.url === 'hello') {
    title = `API Response for /api/v2/hello`;
  }

  return <CodeBlock title={title}>{data && JSON.stringify(data, null, 4)}</CodeBlock>;
}

export default SourceView;
