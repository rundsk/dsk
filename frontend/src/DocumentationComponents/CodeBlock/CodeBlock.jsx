/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';

import { Client } from '@rundsk/js-sdk';
import hljs from 'highlight.js';

import { copyTextToClipboard } from '../../utils';

import './atelier-forest-light.css';
import './CodeBlock.css';

// There are two possible ways this component is used, in the first case
// children is a React object, in the second children is a string with
// (escaped) HTML.
//
// We see the first case when the component is used explictly and the second,
// when it is instantiated by the DocTransformer, turning a <pre> into a
// CodeBlock.
function CodeBlock({ title, src, escaped, language, children, ...props }) {
  const [code, setCode] = useState(null);
  const [copyText, setCopyText] = useState('Copy');
  let isEscaped = !!escaped;
  const codeRef = React.createRef();

  // Sometimes a CodeBlock start with a empty line, because of the way
  // codeblocks have to be formated in Markdown. We consider this undesirable
  // and remove the first line, if it is empty.
  function trimInitialLine(content) {
    if (content.charAt(0) === '\n') {
      return content.substring(1);
    }
    return content;
  }

  function copyCode() {
    setCopyText('Copied!');
    copyTextToClipboard(codeRef.current.textContent);

    setTimeout(() => {
      setCopyText('Copy');
    }, 2000);
  }

  useEffect(() => {
    if (src) {
      Client.fetch(src, 'text').then((data) => setCode(<code className="code-block__code-content">{data}</code>));
    } else {
      let content;

      if (children === undefined) {
        content = '';
      } else if (typeof children === 'string') {
        content = children;
      } else {
        console.debug('Unsupported type (%s) of children content in <CodeBlock>.', typeof children);
        return;
      }
      content = trimInitialLine(content);

      if (isEscaped) {
        setCode(<code className="code-block__code-content" dangerouslySetInnerHTML={{ __html: content }} />);
      } else {
        setCode(<code className="code-block__code-content">{content}</code>);
      }
    }
  }, [src, children, isEscaped]);

  useEffect(() => {
    if (language && codeRef.current) {
      const nodes = codeRef.current.querySelectorAll('code');

      for (let i = 0; i < nodes.length; i++) {
        hljs.highlightElement(nodes[i]);
      }
    }
  }, [code, language, codeRef]);

  let displayTitle = title || src || props['data-node-asset'];

  return (
    <div className="code-block">
      {displayTitle && (
        <div className="code-block__header">
          <div className="code-block__title">{displayTitle}</div>
        </div>
      )}
      <div className="code-block__stage">
        <div className="code-block__copy" onClick={copyCode}>
          {copyText}
        </div>
        <pre className={`code-block__code ${language}`} ref={codeRef}>
          {code}
        </pre>
      </div>
    </div>
  );
}

export default CodeBlock;
