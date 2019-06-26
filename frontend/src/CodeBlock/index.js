import React, { useState } from 'react';
import './CodeBlock.css';
import { copyTextToClipboard } from '../utils';

function CodeBlock(props) {
  let content = props.children;

  // Sometimes a codeblock start with a empty line, because of the way
  // codeblocks have to be formated in Markdown. We consider this
  // undesirable and remove the first line, if it is empty.
  if (content.length === 1 && content[0].charAt(0) === "\n") {
    content = content[0].substring(1);
  }

  const [copyText, setCopyText] = useState("Copy");
  function copyCode() {
    setCopyText("Copied!");
    copyTextToClipboard(content);

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
          {content}
        </code>
      </pre>
    </div>
  );
}

export default CodeBlock;
