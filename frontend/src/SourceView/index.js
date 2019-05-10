import React, { useState, useEffect } from 'react';
import './SourceView.css';
import { Client } from '@atelierdisko/dsk';
import CodeBlock from '../CodeBlock';

function SourceView(props) {
  const [source, setSource] = useState(null);

  useEffect(() => {
    getSource();
  }, [props.url]);

  function getSource() {
    if (props.url !== undefined) {
      Client.get(props.url).then((data) => {
        setSource(data);
      });
    }
  }

  return (
    <CodeBlock title={`API Response for /api/v1/tree/${props.url}`}>
      {source && JSON.stringify(source, null, 4)}
    </CodeBlock>
  );
}

export default SourceView;
