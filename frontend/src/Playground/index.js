/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { Client } from '@rundsk/js-sdk';
import React, { useEffect, useState } from 'react';
import { useGlobal } from 'reactn';

import './ReactPlayground.css';

function Playground(props) {
  const [source] = useGlobal('source');

  const [iframeSourceURL, setIframeSourceURL] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [height, setHeight] = useState("auto");

  const [annotationData, setAnnotationData] = useState({ annotations: [] });
  const [highlightedAnnotation, setHighlightedAnnotation] = useState(null);

  const id = props['data-component'];

  if (props.src) {
    console.warn('<Playground> props.src has been deprecated in favor of props.annotate.');
    props.annotate = props.src;
    delete props.src;
  }

  useEffect(() => {
    Client.playgroundURL(props.node.url, props.doc.id, props['data-component'], source).then(setIframeSourceURL);
  }, [setIframeSourceURL, props.node.url, props.doc.id, props.children, source]);

  console.log(props)

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

  if (props.background === 'checkerboard') {
    classes.push('react-playground--checkerboard');
  }

  if (props.background === 'pinstripes') {
    classes.push('react-playground--pinstripes');
  }

  if (props.background === 'plain') {
    classes.push('react-playground--plain');
  }

  if (props.isPageComponentDemo) {
    classes.push('react-playground--is-page-component-demo');
  }

  let style = {};

  if (props.backgroundcolor) {
    style.backgroundColor = props.backgroundcolor;
  }

  useEffect(() => {
    if (props.annotate) {
      Client.fetch(props.annotate).then(setAnnotationData);
    }
  }, [props.annotate]);

  let annotationMarkers = annotationData.annotations.map((a, i) => {
    let x = `calc(${a.x} + ${a.offsetX ? a.offsetX : '0px'})`;
    let y = `calc(${a.y} + ${a.offsetY ? a.offsetY : '0px'})`;

    return (
      <div
        className={`react-playground__annotation-marker ${
          highlightedAnnotation === i ? 'react-playground__annotation-marker--highlight' : ''
        }`}
        style={{ left: x, top: y }}
        onMouseEnter={() => {
          setHighlightedAnnotation(i);
        }}
        onMouseLeave={() => {
          setHighlightedAnnotation(null);
        }}
        key={i}
      >
        <div
          className="react-playground__annotation-badge react-playground__annotation-badge--highlight"
          style={{ backgroundColor: annotationData.annotationColor }}
        >
          {i + 1}
        </div>
      </div>
    );
  });

  let annotations = annotationData.annotations.map((a, i) => {
    let backgroundColor = highlightedAnnotation === i ? annotationData.annotationColor : '';
    return (
      <div
        className="react-playground__annotation"
        onMouseEnter={() => {
          setHighlightedAnnotation(i);
        }}
        onMouseLeave={() => {
          setHighlightedAnnotation(null);
        }}
        key={i}
      >
        <div
          className={`react-playground__annotation-badge ${
            highlightedAnnotation === i ? 'react-playground__annotation-badge--highlight' : ''
          }`}
          style={{ backgroundColor }}
        >
          {i + 1}
        </div>
        {a.description}
      </div>
    );
  });

  return (
    <div className={classes.join(' ')}>
      <div className="react-playground__stage" style={style}>
        {isLoading && <div className="react-playground__loading-message">Loading Playground …</div>}
        {annotationMarkers}
        <iframe className="react-playground__stage-frame" src={iframeSourceURL}  allowtransparency="true" style={{height}} />
      </div>
      {annotations.length > 0 && <div className="react-playground__annotations">{annotations}</div>}

      {props.caption && <figcaption className="react-playground__caption">{props.caption}</figcaption>}
    </div>
  );
}

export default Playground;
