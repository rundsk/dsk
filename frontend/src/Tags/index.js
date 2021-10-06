/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect, useContext } from 'react';
import contrast from 'get-contrast';

import { GlobalContext } from '../App';

import './Tags.css';

function Tags(props) {
  const [style, setStyle] = useState('');
  const { config, filterTerm, setFilterTerm } = useContext(GlobalContext);

  useEffect(() => {
    if (config?.tags) {
      let styles = config.tags.map((t) => {
        let textColor = 'rgba(0,0,0,.7)';

        if (!t.color.includes('var') && contrast.ratio(t.color, 'black') < contrast.ratio(t.color, 'white')) {
          textColor = 'white';
        }

        if (t.color.includes('--color-red') || t.color.includes('--color-blue')) {
          textColor = 'white';
        }

        return `.tags button.tags__tag[tag*='${t.name}'] { background-color: ${t.color}; color: ${textColor}; }`;
      });
      setStyle(styles.join('\n'));
    }
  }, [config]);

  let tags;

  if (props.tags) {
    tags = props.tags.map((t) => {
      return (
        <li className="tags__tag-item" key={t}>
          <button
            className="tags__tag"
            tag={t}
            onClick={(ev) => {
              if (props.isClickable === false) {
                return;
              }

              if (ev.metaKey) {
                setFilterTerm(`${filterTerm} ${t}`);
              } else {
                setFilterTerm(t);
              }
            }}
          >
            {t}
          </button>
        </li>
      );
    });
  }

  return (
    <ul className="tags">
      <style>{style}</style>

      {tags}
    </ul>
  );
}

export default Tags;
