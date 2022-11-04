/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { useHistory } from 'react-router-dom';

import { constructURL, slugify } from '../../utils';

import './Heading.css';

// props.children[0] has the textContent.
function Heading(props) {
  const history = useHistory();

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

  let id = slugify(getNodeText(props.children));

  let handleClick = (ev) => {
    ev.preventDefault();

    let t = slugify(props.doc.title) + '§' + id;
    constructURL({ activeTab: t });
    history.replace(constructURL({ activeTab: t }));
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

export default Heading;

const getNodeText = (node) => {
  if (['string', 'number'].includes(typeof node)) return node;
  if (node instanceof Array) return node.map(getNodeText).join('');
  if (typeof node === 'object' && node) return getNodeText(node.props.children);
};
