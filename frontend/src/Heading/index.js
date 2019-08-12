/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import { slugify } from '../utils';
import { withRoute } from 'react-router5';

import './Heading.css';

// props.children[0] has the textContent.
function Heading(props) {
  const levels = {
    alpha: 1,
    beta: 2,
    gamma: 3,
    delta: 4,
  };
  const Tag = `h${levels[props.level]}`;

  if (!props.isJumptarget) {
    return <Tag className={`heading heading--${props.level}`}>{props.children}</Tag>;
  }

  let id;

  if (typeof props.children === "object") {
    id = slugify(props.children[0]);
  } else {
    id = slugify(props.children);
  }

  let handleClick = ev => {
    ev.preventDefault();

    let currentRouterState = props.router.getState();
    let currentNode = currentRouterState.params.node || '';
    let t = slugify(props.docTitle) + '§' + id;

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

  // We can’t use just id, because it doesn’t work for number-only headings.
  return (
    <Tag id={id} heading-id={id} className={`heading heading--${props.level}`}>
      <span className="heading__jumplink" onClick={handleClick}>
        §
      </span>
      {props.children}
    </Tag>
  );
}

export default withRoute(Heading);
