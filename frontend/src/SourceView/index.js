/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import './SourceView.css';
import { Client } from '@atelierdisko/dsk';
import CodeBlock from '../CodeBlock';

function SourceView(props) {
  const [source, setSource] = useState(null);

  useEffect(() => {
    if (props.url === "hello") {
      Client.hello().then((data) => {
        setSource(data);
      });
      return;
    }
    getSource();
  }, [props.url]);

  function getSource() {
    if (props.url !== undefined) {
      Client.get(props.url).then((data) => {
        setSource(data);
      });
    }
  }

  let title = `API Response for /api/v2/tree/${props.url}`;

  if (props.url === "hello") {
    title = `API Response for /api/v2/hello`;
  }

  return (
    <CodeBlock title={title}>
      {source && JSON.stringify(source, null, 4)}
    </CodeBlock>
  );
}

export default SourceView;
