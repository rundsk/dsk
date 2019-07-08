/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState } from "react";
import "./AnnotatedImage.css";
import ComponentDemo from "../ComponentDemo";

function AnnotatedImage(props) {
  const [highlightedAnnotation, setHighlightedAnnotation] = useState(null);

  let p = {};
  p.annotations = {
    annnotations: [
      {
        x: "21%",
        y: "8.3%",
        description: "Use a clear label"
      },
      {
        x: "6%",
        y: "41%",
        description: "Pick a color with enough contrast",
        offsetX: "50px"
      }
    ]
  };

  p.src = "/api/v1/tree/docs/The-Frontend/Special-Components/Annotated-Image/list.png";

  return (
    <div className="annotated-image">
      <ComponentDemo>
        <div className="annotated-image__stage">
          <img className="annotated-image__image" src={p.src} alt="" />

          {p.annotations.annnotations.map((a, i) => {
            let x = `calc(${a.x} + ${a.offsetX ? a.offsetX : "0px"})`;
            let y = `calc(${a.y} + ${a.offsetY ? a.offsetY : "0px"})`;

            return (
              <div
                className={`annotated-image__marker ${
                  highlightedAnnotation === i ? "annotated-image__marker--highlight" : ""
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
                  className="annotated-image__badge annotated-image__badge--highlight"
                  style={{ backgroundColor: p.annotations.annotationColor }}
                >
                  {i + 1}
                </div>
              </div>
            );
          })}
        </div>
      </ComponentDemo>

      <div className="annotated-image__annotations">
        {p.annotations.annnotations.map((a, i) => {
          return (
            <div
              className="annotated-image__annotation"
              onMouseEnter={() => {
                setHighlightedAnnotation(i);
              }}
              onMouseLeave={() => {
                setHighlightedAnnotation(null);
              }}
              key={i}
            >
              <div
                className={`annotated-image__badge ${
                  highlightedAnnotation === i ? "annotated-image__badge--highlight" : ""
                }`}
                style={{ backgroundColor: p.annotations.annotationColor }}
              >
                {i + 1}
              </div>
              {a.description}
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default AnnotatedImage;
