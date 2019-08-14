/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState } from 'react';
import './ComponentDemo.css';

function ComponentDemo(props) {
  const [highlightedAnnotation, setHighlightedAnnotation] = useState(null);

  let classes = ['component-demo'];

  if (props.background === 'checkerboard') {
    classes.push('component-demo--checkerboard');
  }

  if (props.background === 'pinstripes') {
    classes.push('component-demo--pinstripes');
  }

  if (props.background === 'plain') {
    classes.push('component-demo--plain');
  }

  let style = {};

  if (props.backgroundcolor) {
    style.backgroundColor = props.backgroundcolor;
  }

  // This should be loaded from a file if there is an annotations pro
  let annotationData = {
    annnotations: [
      {
        x: '53%',
        y: '53%',
        description: 'Use a clear label',
      },
      {
        x: '32%',
        y: '62%',
        description: 'Pick a color with enough contrast',
        offsetX: '50px',
      },
    ],
    annotationColor: '#EE645D',
  };

  if (!props.annotations) {
    annotationData = { annnotations: [] };
  }

  let annotationMarkers = annotationData.annnotations.map((a, i) => {
    let x = `calc(${a.x} + ${a.offsetX ? a.offsetX : '0px'})`;
    let y = `calc(${a.y} + ${a.offsetY ? a.offsetY : '0px'})`;

    return (
      <div
        className={`component-demo__annotation-marker ${
          highlightedAnnotation === i ? 'component-demo__annotation-marker--highlight' : ''
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
          className="component-demo__annotation-badge component-demo__annotation-badge--highlight"
          style={{ backgroundColor: annotationData.annotationColor }}
        >
          {i + 1}
        </div>
      </div>
    );
  });

  let annotations = annotationData.annnotations.map((a, i) => {
    let backgroundColor = highlightedAnnotation === i ? annotationData.annotationColor : '';
    return (
      <div
        className="component-demo__annotation"
        onMouseEnter={() => {
          setHighlightedAnnotation(i);
        }}
        onMouseLeave={() => {
          setHighlightedAnnotation(null);
        }}
        key={i}
      >
        <div
          className={`component-demo__annotation-badge ${
            highlightedAnnotation === i ? 'component-demo__annotation-badge--highlight' : ''
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
      <div className="component-demo__stage" style={style}>
      <div className="component-demo__stage-content">{props.children}</div>
        <div className="component-demo__annotation-marker-stage">{annotationMarkers}</div>
      </div>
      {annotations.length > 0 && <div className="component-demo__annotations">{annotations}</div>}
    </div>
  );
}

export default ComponentDemo;
