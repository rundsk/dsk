import React, { useState } from 'react';
import './CodeBlock.css';
import { copyTextToClipboard } from '../utils';

function CodeBlock(props) {
  const [copyText, setCopyText] = useState("Copy");
  function copyCode() {
    setCopyText("Copied!");
    copyTextToClipboard(props.children);

    setTimeout(() => {
      setCopyText("Copy");
    }, 2000);
  }

  return (
    <div className="code-block">
      <div className="code-block__header">
        <div className="code-block__title">{props.title}</div>
        <div className="code-block__copy" onClick={copyCode}>{copyText}</div>
      </div>
      <pre className="code-block__code">
        <code className="code-block__code-content">
          {props.children}
        </code>
      </pre>
    </div>
  );
}

export default CodeBlock;
