/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect, useContext, useCallback } from 'react';
import { useHistory } from 'react-router-dom';

import { Client } from '@rundsk/js-sdk';

import { constructURL, useURL } from '../utils';
import { GlobalContext } from '../App';

import './SourcePicker.css';

function SourcePicker() {
  const { nodeURL, source: sourceFromURL } = useURL();
  const history = useHistory();

  const [availableSources, setAvailableSources] = useState([]);

  const { source, setSource } = useContext(GlobalContext);

  const switchSource = useCallback(
    (newSource) => {
      setSource(newSource);

      // Update URL
      history.replace(constructURL({ source: newSource }));

      // The node might have gone away.
      Client.has(nodeURL, newSource).then((isExistent) => {
        if (!isExistent) {
          console.log('Current node has gone away after tree has synced.');
          history.push(constructURL({ node: '/', source: newSource }));
        }
      });
    },
    [history, nodeURL, setSource]
  );

  // Handle Socket Event
  useEffect(() => {
    const handleSocketEvent = (ev) => {
      let m = JSON.parse(ev.detail);

      if (m.topic.includes('source.status.changed')) {
        Client.sources().then((data) => {
          setAvailableSources(data.sources);
        });
      }
    };

    window.addEventListener('socketMessage', handleSocketEvent);

    return () => {
      window.removeEventListener('socketMessage', handleSocketEvent);
    };
  }, [setAvailableSources]);

  // Load the available sources
  useEffect(() => {
    console.log('Loading sources');

    Client.sources().then((data) => {
      setAvailableSources(data.sources);
    });
  }, [setAvailableSources]);

  // Set the source
  useEffect(() => {
    if (availableSources.length === 0) {
      return;
    }

    console.log('Setting source');

    let sourceToLoad = null;

    // First we check if the source from the url exists
    if (sourceFromURL && availableSources.some((s) => s.name === sourceFromURL)) {
      sourceToLoad = sourceFromURL;
    }

    if (!sourceToLoad && availableSources.some((s) => s.name === 'live')) {
      sourceToLoad = 'live';
    }

    if (!sourceToLoad) {
      sourceToLoad = availableSources[0].name;
    }

    setSource(sourceToLoad);
  }, [sourceFromURL, availableSources, setSource]);

  return (
    <div className="source-picker">
      <select
        value={source}
        onChange={(ev) => {
          switchSource(ev.target.value);
        }}
      >
        {availableSources.length === 0 && (
          <option disabled selected>
            Loading Versions …
          </option>
        )}
        {availableSources.map((s) => {
          return (
            <option key={s.name} value={s.name} disabled={!s.is_ready}>
              Version: {s.name} {s.is_ready ? '' : '(loading...)'}
            </option>
          );
        })}
      </select>
    </div>
  );
}

export default React.memo(SourcePicker);
