/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect } from 'react';
import { useGlobal } from 'reactn';

import { Client } from '@atelierdisko/dsk';
import { Tree } from '@atelierdisko/dsk';
import { BaseLink, withRoute } from 'react-router5';

import './TreeNavigation.css';

function TreeNavigation(props) {
  const [filterTerm, setFilterTerm] = useGlobal('filterTerm');
  const [filteredTree, setFilteredTree] = useState(null);

  const filterInputRef = React.createRef();

  const shortcutHandler = event => {
    if (event.key === 'Escape') {
      blurFilter();
    }

    if (event.key === 'f' && event.target.nodeName !== 'INPUT') {
      event.preventDefault();
      focusFilter();
    }
  };

  useEffect(() => {
    document.addEventListener('keydown', shortcutHandler);

    return () => {
      document.removeEventListener('keydown', shortcutHandler);
    };
  });

  useEffect(() => {
    filterTree();
  }, [filterTerm]);

  function onFilterTermChange(ev) {
    setFilterTerm(ev.target.value);
  }

  async function filterTree() {
    if (!filterTerm) {
      // No search term given, results in showing the full unfiltered tree (clear).

      setFilteredTree(null);
      return;
    }

    const filter = Client.filter(filterTerm);
    filter.then(data => {
      if (!data.nodes) {
        // Filtering yielded no results, we save us iterating over the
        // existing tree, as we already know what it should look like.
        setFilteredTree(null);
        return;
      }
      let tree = new Tree(props.tree);

      let urls = data.nodes.reduce((carry, node) => {
        carry.push(node.url);
        return carry;
      }, []);

      setFilteredTree(tree.filteredBy(urls).root);
    });
  }

  function renderList(node, activeNode) {
    if (!node) {
      return;
    }

    let classList = ['node'];
    if (activeNode && node.url === activeNode.url) {
      classList.push('node--is-active');
    }

    // let content = [<a href={`/tree/${node.url}`} key={"link"}>{node.title}</a>]
    let content = [
      <BaseLink
        router={props.router}
        routeName="node"
        routeParams={{ node: `${node.url}` }}
        key={'link'}
        onClick={props.hideMobileSidebar}
      >
        {node.title}
      </BaseLink>,
    ];

    let children = node.children.map(c => {
      return renderList(c, activeNode);
    });

    if (children) {
      content.push(<ul key={'children'}>{children}</ul>);
    }

    return <li key={node.title}>{content}</li>;
  }

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

  let tree = renderList(filteredTree || props.tree);

  // We throw out the root node and only display
  // the children
  if (tree && tree.props.children) {
    tree = tree.props.children[1];
  }

  if (filteredTree && filteredTree.children.length === 0) {
    tree = <div className="tree-navigation__empty">No aspects found</div>;
  }

  return (
    <nav className="tree-navigation">
      <div className="tree-navigation__tree">{tree}</div>

      <div className="tree-navigation__filter">
        <input
          type="search"
          placeholder="Filter Aspects"
          value={filterTerm}
          onChange={onFilterTermChange}
          ref={filterInputRef}
        />
      </div>
    </nav>
  );
}

export default withRoute(TreeNavigation);
