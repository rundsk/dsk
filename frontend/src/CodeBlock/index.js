/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState } from 'react';
import './CodeBlock.css';
import { copyTextToClipboard } from '../utils';

// TODO: If child is <code> unwrap and turn into string.
function CodeBlock(props) {
  let content = props.children;

  // Sometimes a codeblock start with a empty line, because of the way
  // codeblocks have to be formated in Markdown. We consider this
  // undesirable and remove the first line, if it is empty.
  if (
    React.Children.count(props.children) === 1 &&
    typeof props.children[0] === 'string' &&
    props.children[0].charAt(0) === '\n'
  ) {
    content = props.children[0].substring(1);
  }

  const [copyText, setCopyText] = useState('Copy');

  function copyCode() {
    setCopyText('Copied!');
    copyTextToClipboard(content);

    setTimeout(() => {
      setCopyText('Copy');
    }, 2000);
  }

  let code;
  if (props.escaped) {
    code = <code className="code-block__code-content" dangerouslySetInnerHTML={{__html: content}} />;
  } else {
    code = <code className="code-block__code-content">{content}</code>
  }

  return (
    <div className="code-block">
      {props.title && (
        <div className="code-block__header">
          <div className="code-block__title">{props.title}</div>
        </div>
      )}
      <div className="code-block__stage">
        <div className="code-block__copy" onClick={copyCode}>
          {copyText}
        </div>
        <pre className="code-block__code">
          {code}
        </pre>
      </div>
    </div>
  );
}

export default CodeBlock;
