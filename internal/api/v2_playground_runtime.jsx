/**
 * Copyright 2021 Marius Wilms, Christoph Labacher. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useEffect, useLayoutEffect } from "react";
import ReactDOM from "react-dom";

const handleOnLoad = () => {
  const id = document.querySelector("body").getAttribute("data-id");
  window.parent.postMessage(
    JSON.stringify({
      id,
      contentHeight: document.querySelector("html").offsetHeight,
    }),
    "*"
  );
};

const PlaygroundWrapper = () => {
  useLayoutEffect(handleOnLoad);

  useEffect(() => {
    // This is called after all images loaded
    window.addEventListener("load", handleOnLoad);

    return () => {
      window.removeEventListener("load", handleOnLoad);
    };
  }, []);

  useEffect(() => {
    const resizeObserver = new ResizeObserver(handleOnLoad);
    resizeObserver.observe(document.body);

    return () => {
      resizeObserver.unobserve(document.body);
    }
  }, []);

  const noPadding = window.frameElement.attributes.nopadding;

  /* eslint-disable react/jsx-no-undef */
  return (
    <div
      style={{
        width: "100%",
        paddingTop: noPadding ? 0 : 48,
        paddingBottom: noPadding ? 0 : 48,
        paddingLeft: noPadding ? 0 : 64,
        paddingRight: noPadding ? 0 : 64,
      }}
    >
      <ThePlaygroundInQuestion />
    </div>
  );
  /* eslint-enable */
};

document.addEventListener("DOMContentLoaded", () => {
  ReactDOM.render(<PlaygroundWrapper />, document.getElementById("root"));
});
