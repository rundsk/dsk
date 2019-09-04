/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import { useGlobal } from 'reactn';
import './FigmaEmbed.css';

function FigmaEmbed(props) {
  const [config, setConfig] = useGlobal('config');
  const [image, setImage] = useState(null);
  const [frameId, setFrameId] = useState(null);
  const [errorMessage, setErrorMessage] = useState(null);
  const [loadingMessage, setLoadingMessage] = useState(null);

  // Retrieves document.
  useEffect(() => {
    if (!config.figma.accessToken) {
      setErrorMessage(
        'Missing personal access token, please visit: https://rundsk.com/tree/The-Frontend/Configuration'
      );
      return;
    }
    if (!props.document) {
      setErrorMessage('No document given.');
      return;
    }
    setLoadingMessage('Loading document …');

    fetch(`https://api.figma.com/v1/files/${props.document}`, {
      method: 'GET',
      headers: new Headers({
        'X-Figma-Token': config.figma.accessToken,
      }),
    })
      .then(response => {
        if (response.status === 200) {
          return response.json();
        } else {
          setErrorMessage('Something went wrong.');
        }
      })
      .then(findId)
      .catch(err => {
        console.log(err);
        setErrorMessage('Something went wrong.');
      });
  }, [props.document, props.frame, config]);

  function findId(data) {
    let nameWeAreLookingFor = props.frame;
    let nodeId = undefined;

    let filter = node => {
      if (node.name === nameWeAreLookingFor) {
        nodeId = node.id;
      } else {
        if (node.children && node.children.length > 0) {
          node.children.forEach(filter);
        }
      }
    };

    filter(data.document);

    if (nodeId === undefined) {
      setErrorMessage('No frame with the specified title was found in your document.');
      return;
    }

    getImage(nodeId);
  }

  function getImage(nodeId) {
    if (nodeId && props.document && config.figma.accessToken) {
      setLoadingMessage(`Loading image for “${props.frame}” …`);

      return fetch(`https://api.figma.com/v1/images/${props.document}?ids=${nodeId}`, {
        method: 'GET',
        headers: new Headers({
          'X-Figma-Token': config.figma.accessToken,
        }),
      })
        .then(response => {
          if (response.status === 200) {
            return response.json();
          } else {
            setErrorMessage('Something went wrong.');
          }
        })
        .then(data => {
          setImage(data.images[nodeId]);
          setFrameId(nodeId);
        })
        .catch(err => {
          console.log(err);
          setErrorMessage('Something went wrong.');
        });
    }
  }

  return (
    <div className="figma-embed">
      {image && <img src={image} alt={props.frame} />}
      {!image && !errorMessage && (
        <div className="figma-embed__loader">
          {loadingMessage ? loadingMessage : `Loading “${props.frame}” from Figma…`}
        </div>
      )}
      {errorMessage && <div className="figma-embed__error">{errorMessage}</div>}

      <a
        className="figma-embed__via"
        href={`https://www.figma.com/file/${props.document}${frameId ? `?node-id=${frameId}` : ''}`}
        target="_blank"
        rel="noopener noreferrer"
      >
        Edit in Figma
      </a>
    </div>
  );
}

export default FigmaEmbed;
