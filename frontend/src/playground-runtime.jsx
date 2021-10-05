import React, { useLayoutEffect } from 'react';
import ReactDOM from 'react-dom';

const PlaygroundWrapper = () => {
  useLayoutEffect(() => {
    const id = document.querySelector("body").getAttribute("data-id");
    window.parent.postMessage(JSON.stringify({
      id,
      contentHeight: document.querySelector("#root").clientHeight,
    }), '*');
  })

  return <div style={{
    width: "100%",
    padding: 40
  }}><ThePlaygroundInQuestion /></div>
}

document.addEventListener('DOMContentLoaded', () => {
  ReactDOM.render(<PlaygroundWrapper />, document.getElementById('root'));
});
