/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import './TableOfContents.css';
import { Client } from '@rundsk/js-sdk';
import { slugify } from '../utils';
import { withRoute } from 'react-router5';

function TOCEntry(props) {
  if (props.cutoffLevel && props.level >= props.cutoffLevel) {
    return <></>;
  }

  let slug = slugify(props.title);

  let children = [];
  if (props.children) {
    children = props.children.map(c => (
      <TOCEntry {...c} docTitle={props.docTitle} onClick={props.onClick} cutoffLevel={props.cutoffLevel} />
    ));
  }

  return (
    <li>
      <a
        href={slugify(props.docTitle) + '§' + slug}
        onClick={ev => {
          props.onClick(ev, slug);
        }}
      >
        {props.title}
      </a>
      {children.length > 0 && <ul>{children}</ul>}
    </li>
  );
}

function TableOfContents(props) {
  const [data, setData] = useState([]);

  useEffect(() => {
    let node = props.src;
    if (!node) {
      let currentRouterState = props.router.getState();
      node = currentRouterState.params.node || '';
    }

    Client.get(node).then(data => {
      let doc = data.docs.find(d => d.title === props.docTitle);
      if (doc && doc.toc) {
        setData(doc.toc);
      }
    });
  }, [props.src, props.router, props.docTitle]);

  let handleClick = (ev, slug) => {
    ev.preventDefault();

    let currentRouterState = props.router.getState();
    let currentNode = currentRouterState.params.node || '';
    let t = slugify(props.docTitle) + '§' + slug;

    props.router.navigate(
      'node',
      {
        ...currentRouterState.params,
        node: currentNode,
        t,
      },
      { replace: true }
    );
  };

  return (
    <nav className="table-of-contents">
      <div className="table-of-contents__title">Contents</div>
      <ul>
        {data.map(e => {
          return <TOCEntry {...e} docTitle={props.docTitle} onClick={handleClick} cutoffLevel={props.cutofflevel} />;
        })}
      </ul>
    </nav>
  );
}

export default withRoute(TableOfContents);
