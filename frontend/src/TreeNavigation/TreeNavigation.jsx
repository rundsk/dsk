/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect, useContext, useMemo, useLayoutEffect } from 'react';

import { Client } from '@rundsk/js-sdk';
import { Tree } from '@rundsk/js-sdk';
import { NavLink } from '../Link';

import { GlobalContext } from '../App';

import './TreeNavigation.css';

function renderList(node, onHideMobileSidebar) {
  if (!node) {
    return;
  }

  let content = [
    <NavLink activeClassName={'is-active'} exact to={`/${node.url}`} key={'Navlink'} onClick={onHideMobileSidebar}>
      {node.title}
    </NavLink>,
  ];

  let children = node.children.map((c) => {
    return renderList(c, onHideMobileSidebar);
  });

  if (children) {
    content.push(<ul key={'children'}>{children}</ul>);
  }

  return <li key={node.title}>{content}</li>;
}

function TreeNavigation({ tree, onHideMobileSidebar }) {
  const { filterTerm, setFilterTerm } = useContext(GlobalContext);

  const [filteredTree, setFilteredTree] = useState(null);

  const filterInputRef = React.createRef();

  const shortcutHandler = (event) => {
    if (event.key === 'Escape') {
      blurFilter();
    }

    if (event.key === 'f' && event.target.nodeName !== 'INPUT') {
      event.preventDefault();
      focusFilter();
    }
  };

  function blurFilter() {
    if (filterInputRef.current) {
      filterInputRef.current.blur();
    }
  }

  function focusFilter() {
    if (filterInputRef.current) {
      filterInputRef.current.focus();
    }
  }

  useEffect(() => {
    document.addEventListener('keydown', shortcutHandler);

    return () => {
      document.removeEventListener('keydown', shortcutHandler);
    };
  });

  function handleFilterTermChange(ev) {
    setFilterTerm(ev.target.value);
  }

  // Scroll the link to the currently open page into view on page load
  useLayoutEffect(() => {
    const activeLink = document.querySelector('.tree-navigation__tree .is-active');

    if (activeLink) {
      activeLink.scrollIntoView();
    }
  });

  useEffect(() => {
    async function filterTree() {
      if (!filterTerm) {
        // No search term given, results in showing the full unfiltered tree (clear).
        setFilteredTree(null);
        return;
      }

      Client.filter(filterTerm)
        .then((data) => {
          if (!data.nodes) {
            // Filtering yielded no results, we save us iterating over the
            // existing tree, as we already know what it should look like.
            setFilteredTree(null);
            return;
          }
          let treeToFilter = new Tree(tree);

          setFilteredTree(treeToFilter.filteredBy(data.nodes.map((n) => n.url)).root);
        })
        .catch((error) => {
          console.log(error);
        });
    }

    filterTree();
  }, [filterTerm, tree]);

  let renderedTree = useMemo(() => {
    return renderList(filteredTree || tree, onHideMobileSidebar);
  }, [tree, filteredTree, onHideMobileSidebar]);

  // We throw out the root node and only display
  // the children
  if (renderedTree && renderedTree.props.children) {
    renderedTree = renderedTree.props.children[1];
  }

  if (filteredTree && filteredTree.children.length === 0) {
    renderedTree = <div className="tree-navigation__empty">No aspects found</div>;
  }

  return (
    <nav className="tree-navigation">
      <div className="tree-navigation__tree">{renderedTree}</div>

      <div className="tree-navigation__filter">
        <input
          type="search"
          placeholder="Filter Aspects"
          value={filterTerm}
          onChange={handleFilterTermChange}
          ref={filterInputRef}
        />
      </div>
    </nav>
  );
}

export default React.memo(TreeNavigation);
