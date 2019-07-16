/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import './FigmaEmbed.css';

// FIXME: This component should not use my persolan
// access token
function FigmaEmbed(props) {
  const [image, setImage] = useState(null);
  const [frameId, setFrameId] = useState(null);
  const [errorMessage, setErrorMessage] = useState(null);
  const [loadingMessage, setLoadingMessage] = useState(null);

  useEffect(() => {
    getDocument();
  }, [props.document, props.frame]);

  function getDocument() {
    if (props.document && props.token) {
      const myHeaders = new Headers();
      myHeaders.append('X-Figma-Token', props.token);

      setLoadingMessage('Loading document …');

      fetch(`https://api.figma.com/v1/files/${props.document}`, {
        method: 'GET',
        headers: myHeaders,
      })
        .then((response) => {
          if (response.status === 200) {
            return response.json();
          } else {
            setErrorMessage('Something went wrong.');
          }
        })
        .then((data) => {
          findId(data);
        })
        .catch((err) => {
          console.log(err);
          setErrorMessage('Something went wrong.');
        });

      // fetch(`https://api.figma.com/v1/teams/669136806690498635/styles`, {
      //   method: 'GET',
      //   headers: myHeaders,
      // }).then((response) => {
      //   if (response.status === 200) {
      //     return response.json();
      //   } else {
      //     throw new Error('Something went wrong on api server!');
      //   }
      // }).then((data) => {
      //   console.log(data)
      // }).catch((err) => {
      //   console.log(err);
      // });
    }
  }

  function findId(data) {
    let nameWeAreLookingFor = props.frame;
    let nodeId = undefined;

    let filter = (node) => {
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
    if (nodeId && props.document && props.token) {
      setLoadingMessage(`Loading image for “${props.frame}” …`);

      const myHeaders = new Headers();
      myHeaders.append('X-Figma-Token', props.token);
      return fetch(`https://api.figma.com/v1/images/${props.document}?ids=${nodeId}`, {
        method: 'GET',
        headers: myHeaders,
      })
        .then((response) => {
          if (response.status === 200) {
            return response.json();
          } else {
            setErrorMessage('Something went wrong.');
          }
        })
        .then((data) => {
          setImage(data.images[nodeId]);
          setFrameId(nodeId);
        })
        .catch((err) => {
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
