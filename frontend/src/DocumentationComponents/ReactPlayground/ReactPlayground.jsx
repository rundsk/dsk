/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { Client } from '@rundsk/js-sdk';
import React, { useContext, useEffect, useState } from 'react';

import { GlobalContext } from '../../App';
import Playground from '../Playground';
import CodeBlock from '../CodeBlock';

import './ReactPlayground.css';

function ReactPlayground(props) {
  const { source } = useContext(GlobalContext);

  const [showPlaygroundSource, setShowPlaygroundSource] = useState(props.showsource);

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
    <div className={`react-playground ${showPlaygroundSource && 'react-playground--show-source'}`}>
      <div className="react-playground__content">
        <Playground {...props} caption={null} noPadding contentFullWidth>
          {isLoading && <div className="react-playground__loading-message">Loading Playground …</div>}
          <iframe
            className="react-playground__stage-frame"
            src={iframeSourceURL}
            allowtransparency="true"
            style={{ height: isLoading ? 0 : height }}
            title={id}
            nopadding={props.nopadding}
          />
        </Playground>

        {!props.showsource && !props.disableshowsource && (
          <button
            className="react-playground__show-source"
            onClick={() => {
              setShowPlaygroundSource(!showPlaygroundSource);
            }}
            type="button"
            style={{ backgroundColor: props.nopadding ? 'transparent' : props.backgroundcolor }}
          >
            {showPlaygroundSource ? 'Hide' : 'Show'} Source
          </button>
        )}
      </div>
      {showPlaygroundSource && (
        <CodeBlock language="jsx" escaped>
          {props.children}
        </CodeBlock>
      )}

      {props.caption && <figcaption className="playground__caption">{props.caption}</figcaption>}
    </div>
  );
}

export default React.memo(ReactPlayground);
