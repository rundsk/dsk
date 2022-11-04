/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { useHistory } from 'react-router-dom';

import { constructURL, slugify } from '../../utils';

import './TableOfContents.css';

function TOCEntry(props) {
  if (props.cutoffLevel && props.level >= props.cutoffLevel) {
    return <></>;
  }

  let slug = slugify(props.title);

  let children = [];
  if (props.children) {
    children = props.children.map((c) => (
      <TOCEntry {...c} key={c.title} doc={props.doc} onClick={props.onClick} cutoffLevel={props.cutoffLevel} />
    ));
  }

  return (
    <li>
      <a
        href={slugify(props.doc.title) + '§' + slug}
        onClick={(ev) => {
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
  const history = useHistory();

  let handleClick = (ev, slug) => {
    ev.preventDefault();

    let t = slugify(props.doc.title) + '§' + slug;
    history.replace(constructURL({ activeTab: t }));
  };

  return (
    <nav className="table-of-contents">
      <div className="table-of-contents__title">Contents</div>
      <ul>
        {props.doc.toc.map((e) => {
          return (
            <TOCEntry {...e} key={e.title} doc={props.doc} onClick={handleClick} cutoffLevel={props.cutofflevel} />
          );
        })}
      </ul>
    </nav>
  );
}

export default TableOfContents;
