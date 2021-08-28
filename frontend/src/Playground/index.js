/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { Client } from '@rundsk/js-sdk';
import React, { useEffect, useState } from 'react';
import ReactDOMServer from 'react-dom/server';
import { useGlobal } from 'reactn';
import './Playground.css';

function Playground(props) {
  const [source] = useGlobal('source');

  const [iframeSourceURL, setIframeSourceURL] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  const [annotationData, setAnnotationData] = useState({ annotations: [] });
  const [highlightedAnnotation, setHighlightedAnnotation] = useState(null);

  if (props.src) {
    console.warn('<Playground> props.src has been deprecated in favor of props.annotate.');
    props.annotate = props.src;
    delete props.src;
  }

  useEffect(() => {
    Client.playgroundURL(props.node.url, props.doc.id, props['data-component'], source).then(setIframeSourceURL);
  }, [setIframeSourceURL, props.node.url, props.doc.id, props.children, source]);

  // Handle communication with iframe. This is currently happens only from the iframe to us.
  useEffect(() => {
    const handler = (ev) => {
      let data = JSON.parse(ev.data);

      console.debug('Received message from playground iframe:', data);

      if (data?.status === 'ready') {
        setIsLoading(false);
      }
    };
    window.addEventListener('message', handler);
    return () => window.removeEventListener('message', handler);
  }, [setIsLoading]);

  let classes = ['playground'];

  if (isLoading) {
    classes.push('is-loading');
  }

  if (props.background === 'checkerboard') {
    classes.push('playground--checkerboard');
  }

  if (props.background === 'pinstripes') {
    classes.push('playground--pinstripes');
  }

  if (props.background === 'plain') {
    classes.push('playground--plain');
  }

  if (props.isPageComponentDemo) {
    classes.push('playground--is-page-component-demo');
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
        className={`playground__annotation-marker ${
          highlightedAnnotation === i ? 'playground__annotation-marker--highlight' : ''
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
          className="playground__annotation-badge playground__annotation-badge--highlight"
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
        className="playground__annotation"
        onMouseEnter={() => {
          setHighlightedAnnotation(i);
        }}
        onMouseLeave={() => {
          setHighlightedAnnotation(null);
        }}
        key={i}
      >
        <div
          className={`playground__annotation-badge ${
            highlightedAnnotation === i ? 'playground__annotation-badge--highlight' : ''
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
      <div className="playground__stage" style={style}>
        {isLoading && <div className="playground__loading-message">Loading Playground …</div>}
        {annotationMarkers}
        <iframe className="playground__stage-frame" src={iframeSourceURL} />
      </div>
      {annotations.length > 0 && <div className="playground__annotations">{annotations}</div>}

      {props.caption && <figcaption className="playground__caption">{props.caption}</figcaption>}
    </div>
  );
}

export default Playground;
