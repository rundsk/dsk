/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import './Playground.css';
import { Client } from '@atelierdisko/dsk';

function Playground(props) {
  const [annotationData, setAnnotationData] = useState({ annotations: [] });
  const [highlightedAnnotation, setHighlightedAnnotation] = useState(null);

  let classes = ['playground'];

  if (props.background === 'checkerboard') {
    classes.push('playground--checkerboard');
  }

  if (props.background === 'pinstripes') {
    classes.push('playground--pinstripes');
  }

  if (props.background === 'plain') {
    classes.push('playground--plain');
  }

  let style = {};

  if (props.backgroundcolor) {
    style.backgroundColor = props.backgroundcolor;
  }

  useEffect(() => {
    if (props.src) {
      Client.fetch(props.src).then(data => setAnnotationData(data));
    }
  }, [props.src]);

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
        <div className="playground__stage-content">{props.children}</div>
        <div className="playground__annotation-marker-stage">{annotationMarkers}</div>
      </div>
      {annotations.length > 0 && <div className="playground__annotations">{annotations}</div>}
    </div>
  );
}

export default Playground;
