/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';

import { Client } from '@rundsk/js-sdk';

import hljs from 'highlight.js/lib/core';
import javascript from 'highlight.js/lib/languages/javascript';
import css from 'highlight.js/lib/languages/css';
import bash from 'highlight.js/lib/languages/bash';
import php from 'highlight.js/lib/languages/php';
import go from 'highlight.js/lib/languages/go';
import nginx from 'highlight.js/lib/languages/nginx';
import dockerfile from 'highlight.js/lib/languages/dockerfile';

import { copyTextToClipboard } from '../../utils';

import './atelier-forest-light.css';
import './CodeBlock.css';

hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('css', css);
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('php', php);
hljs.registerLanguage('go', go);
hljs.registerLanguage('nginx', nginx);
hljs.registerLanguage('dockerfile', dockerfile);

// This components receives a literal string with HTML escaped content as its
// children. Independent if indirectly used with <pre> tags (coming from fenced
// Markdown syntax) or directly when <CodeBlock> was used inside the document.
//
// The DocTransformer will ensure that the content is considered pre-formatted
// in any case.
function CodeBlock({ title, src, language, children, ...props }) {
  const [code, setCode] = useState(null);
  const [copyText, setCopyText] = useState('Copy');
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

      setCode(<code className="code-block__code-content" dangerouslySetInnerHTML={{ __html: content }} />);
    }
  }, [src, children]);

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
        <pre className={`code-block__code ${language || ''}`} ref={codeRef}>
          {code}
        </pre>
      </div>
    </div>
  );
}

export default CodeBlock;
