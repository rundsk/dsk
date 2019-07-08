/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useEffect } from 'react';
import './ComponentDemo.css';

function ComponentDemo(props) {
  const ref = React.createRef();

  useEffect(() => {
    setRetinaImageSize();
  });

  function setRetinaImageSize() {
    if (ref.current) {
      let node = ref.current;

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
      let imgs = node.querySelectorAll("img");
      imgs.forEach(img => {
        let src = img.getAttribute("src");

        if (src.includes("@2x")) {
          let width = img.getAttribute("width");
          let height = img.getAttribute("height");

          img.style.maxWidth = `${width/2}px`;
          img.style.maxHeight = `${height/2}px`;
        }
      });
    }
  }


    let classes = ['component-demo'];

    if (props.background === "checkerboard") {
      classes.push('component-demo--checkerboard');
    }

    if (props.background === "pinstripes") {
      classes.push('component-demo--pinstripes');
    }

    if (props.background === "plain") {
      classes.push('component-demo--plain');
    }

    let style = {};

    if (props.backgroundcolor) {
      style.backgroundColor = props.backgroundcolor;
    }

    return (
      <div className={classes.join(' ')} ref={ref} style={style}>
        {props.children}
      </div>
    );
}

export default ComponentDemo;
