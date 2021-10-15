/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useContext, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';

import { Client } from '@rundsk/js-sdk';

import { constructURL, useURL } from '../utils';
import { GlobalContext } from '../App';
import ErrorPage from '../ErrorPage';
import Page from '../Page';

const Node = React.memo(({ nodeURL, source, activeTab }) => {
  const history = useHistory();
  const [nodeData, setNodeData] = useState(null);
  const [error, setError] = useState(null);

  // Handle Socket Event
  useEffect(() => {
    const handleSocketEvent = (ev) => {
      let m = JSON.parse(ev.detail);

      if (m.topic === `${source}.tree.synced`) {
        // The node might have gone away.
        Client.has(nodeURL, source).then((isExistent) => {
          if (isExistent) {
            loadNode(nodeURL, source);
          } else {
            console.log('Current node has gone away after tree has synced.');
            history.push(constructURL({ node: '/', source: source }));
          }
        });
      }
    };

    window.addEventListener('socketMessage', handleSocketEvent);

    return () => {
      window.removeEventListener('socketMessage', handleSocketEvent);
    };
  }, [history, nodeURL, source]);

  function loadNode(nodeURL, source) {
    if (!source) {
      return;
    }
    Client.get(nodeURL, source)
      .then((data) => {
        setNodeData({ ...data, source: source });
        setError(null);
      })
      .catch((err) => {
        console.log(`Failed to set node data: ${err}`);
        setError('Design aspect not found.');
      });
  }

  useEffect(() => {
    loadNode(nodeURL, source);
  }, [nodeURL, source]);

  if (error) {
    return <ErrorPage>{error}</ErrorPage>;
  }

  if (nodeData) {
    return <Page {...nodeData} activeTab={activeTab} />;
  }

  return <div />;
});

const NodeWrapper = () => {
  const { nodeURL, activeTab } = useURL();
  const { source } = useContext(GlobalContext);

  return <Node nodeURL={nodeURL || ''} source={source} activeTab={activeTab} />;
};

export default NodeWrapper;
