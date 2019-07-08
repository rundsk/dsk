/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from "react";
import "./ColorSpecimen.css";
import { Client } from "@atelierdisko/dsk";
import { copyTextToClipboard } from "../utils";
import contrast from "get-contrast";

function ColorSpecimen(props) {
  const [colors, setColors] = useState([]);

  useEffect(() => {
    getData();
  }, [props.src]);

  function getData() {
    if (props.src) {
      // FIXME: This should obviously be derived from the src attribute
      Client.get(`/Basics/Colors${props.src.slice(1)}`).then(data => {
        setColors(data.colors);
        console.log(data.colors);
      });
    }
  }

  let classes = ["color-specimen"];

  if (props.compact) {
    classes.push("color-specimen--is-compact");
  }

  return (
    <div className={classes.join(" ")}>
      {colors.map(c => {
        let classes = ["color-sample"];

        if (contrast.ratio(c.value, "white") < 1.5) {
          classes.push("color-sample--is-ultra-light");
        }

        return (
          <div
            className={classes.join(" ")}
            key={c.id}
            onClick={() => {
              copyTextToClipboard(c.value);
            }}
          >
            <div className="color-sample__demo" style={{ backgroundColor: c.value }}>
              <div className="color-sample__score">
                <span>{contrast.score(c.value, "white")}</span>
                <span>{contrast.score(c.value, "black")}</span>
              </div>
            </div>
            <div className="color-sample__name">
              {c.name} <span className="color-sample__id">({c.id})</span>
            </div>
            <div className="color-sample__spec">{c.value}</div>
            <div className="color-sample__comment">{c.comment}</div>
          </div>
        );
      })}
    </div>
  );
}

export default ColorSpecimen;
