/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState } from 'react';
import ReactDOMServer from 'react-dom/server';
import './CodeBlock.css';
import { copyTextToClipboard } from '../utils';

// There are two possible ways this component is used, in the first case
// children is a React object, in the second children is a string with
// (escaped) HTML.
//
// We see the first case when the component is used explictly and the second,
// when it is instantiated by the DocTransformer, turning a <pre> into a
// CodeBlock.
function CodeBlock(props) {
  let isEscaped = props.escaped;

  // Sometimes a CodeBlock start with a empty line, because of the way
  // codeblocks have to be formated in Markdown. We consider this undesirable
  // and remove the first line, if it is empty.
  function trimInitialLine(content) {
    if (content.charAt(0) === '\n') {
      return content.substring(1);
    }
    return content;
  }

  const [copyText, setCopyText] = useState('Copy');

  let content;
  if (props.children === undefined) {
    content = '';
  } else if (typeof props.children === 'object') {
    content = ReactDOMServer.renderToStaticMarkup(props.children);
  } else if (typeof props.children === 'string') {
    content = props.children;
  }
  content = trimInitialLine(content);

  function copyCode() {
    setCopyText('Copied!');
    copyTextToClipboard(content);

    setTimeout(() => {
      setCopyText('Copy');
    }, 2000);
  }

  let code;
  if (isEscaped) {
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
