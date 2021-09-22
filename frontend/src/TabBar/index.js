/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import './Tabs.css';
import { slugify } from '../utils';

export function Tab() {
  return <div></div>;
}

function TabBar(props) {
  let tabs = [];

  tabs =
    props.tabs &&
    props.tabs.map((t, i) => {
      let classes = ['tab-bar__tab'];

      // If no active tab is set, the first on is considered active
      if (props.activeTab === undefined && i === 0) {
        classes.push('tab-bar__tab--is-active');
      }

      if (props.activeTab === slugify(t)) {
        classes.push('tab-bar__tab--is-active');
      }

      return (
        <a
          href={`#${t}`}
          className={classes.join(' ')}
          key={t}
          onClick={(ev) => {
            ev.preventDefault();
            props.onSetActiveTab(t);
          }}
        >
          {t}
        </a>
      );
    });

  let rightSideTabs = [];

  rightSideTabs =
    props.rightSideTabs &&
    props.rightSideTabs.map((t, i) => {
      let classes = ['tab-bar__tab'];

      if (props.activeTab === slugify(t)) {
        classes.push('tab-bar__tab--is-active');
      }

      return (
        <a
          href={`#${t}`}
          className={classes.join(' ')}
          key={t}
          onClick={(ev) => {
            ev.preventDefault();
            props.onSetActiveTab(t);
          }}
        >
          {t}
        </a>
      );
    });

  return (
    <div className="tab-bar">
      <div>{tabs}</div>
      <div>{rightSideTabs}</div>
    </div>
  );
}

export default TabBar;
