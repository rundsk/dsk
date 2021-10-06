/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { Client } from '@rundsk/js-sdk';
import React, { useContext, useEffect, useState } from 'react';

import { GlobalContext } from '../../App';
import Playground from '../Playground';

import './ReactPlayground.css';

function ReactPlayground(props) {
  const { source } = useContext(GlobalContext);

  const [iframeSourceURL, setIframeSourceURL] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [height, setHeight] = useState('auto');

  const id = props['data-component'];

  useEffect(() => {
    Client.playgroundURL(props.node.url, props.doc.id, id, source).then(setIframeSourceURL);
  }, [setIframeSourceURL, props.node.url, props.doc.id, props.children, id, source]);

  // Handle communication with iframe. This is currently happens only from the iframe to us.
  useEffect(() => {
    const handler = (ev) => {
      let data = JSON.parse(ev.data);

      console.debug('Received message from playground iframe:', data);

      if (!data || data?.id !== id) {
        return;
      }

      if (data.status === 'ready') {
        setIsLoading(false);
      }

      if (data.contentHeight) {
        setHeight(data.contentHeight);
      }
    };
    window.addEventListener('message', handler);
    return () => window.removeEventListener('message', handler);
  }, [setIsLoading, setHeight, id]);

  let classes = ['react-playground'];

  if (isLoading) {
    classes.push('react-playground--is-loading');
  }

  return (
    <Playground {...props} noPadding contentFullWidth>
      {isLoading && <div className="react-playground__loading-message">Loading Playground …</div>}
      <iframe
        className="react-playground__stage-frame"
        src={iframeSourceURL}
        allowtransparency="true"
        style={{ height: isLoading ? 0 : height }}
        title={id}
      />
    </Playground>
  );
}

export default React.memo(ReactPlayground);
