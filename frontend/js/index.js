/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Tree: false */
/* globals URLSearchParams: false */

"use strict";

document.addEventListener('DOMContentLoaded', function() {
  const $ = document.querySelectorAll.bind(document);
  const $1 = document.querySelector.bind(document);

  let header = $1('header');
  let main = $1('main');
  let nav = $1('.tree-nav');
  let searchField = $1('.search__field');
  let searchClear = $1('.search__clear');

  let tree = new Tree();
  let baseTitle = document.title;

  // Initializes the tree, the left hand tree navigation and displays
  // the currently route selected node - if any.
  tree.sync()
    .then(() => {
      return tree.search();
    })
    .then((results) => {
      // navigateToNode(window.location.pathname, true);
      // handleSearchWithQuery(window.location.search.substring(1));
    });


  // main.querySelectorAll('.text a, .crumbs a, .keywords a, .children-table a').forEach((el) => {
  main.addEvenListener('click', (ev) => {
    if (ev.target && ev.target.nodeName == 'A') {
      console.log('Clicked link in main!', ev.target);
      /// el.addEventListener('click', handleTextLinkClick);
    }
  });
  header.querySelector('a').addEventListener('click', handleTextLinkClick);

  // Load the node based on node URL path.
  function navigateToNode(url, initialNavigation) {
    return tree.get(url).then((node) => {
      let params = new URLSearchParams(window.location.search);

      history[initialNavigation ? 'replaceState' : 'pushState'](
          {node: node, search: params.get('q')},
          '' ,
          '/' + url + '?q=' + params.get('q')
        );

      document.title = baseTitle + ': ' + node.title;
      main.innerHTML = renderNode(node);
      activateNode(node);
    });
  }

  //
  // HTML Rendering
  //

  // TODO Implement.
  function renderNode(node) {
    return node.description;
  }

  // Renders the nav structure.
  // FIXME: Append list withouth root node?
  // if list nav.appendChild(list.childNodes[1]);
  function renderNav(tree) {
    return renderList(tree.root);
  }

  // Recursively turns the given tree data into a "ul li" structure.
  function renderList(node) {
    let ul = document.createElement('ul');
    let li = document.createElement('li');
    let a  = document.createElement('a');

    a.href = "/" + node.url;
    a.innerHTML = node.title;
    a.addEventListener('click', handleNav);

    if (node.isGhost) {
      a.classList.add('ghosted');
    }
    li.appendChild(a);
    li.appendChild(ul);

    for (let child in node.children) {
      let childList = renderList(node.children[child]);
      if (childList) {
        ul.appendChild(childList);
      }
    }
    return li;
  }

  function activateNode(node) {
    let bullets = nav.querySelectorAll('li');

    for (let a of bullets) {
      a.classList.remove('is-active');
    }

    let activeNode = bullets.querySelector('a[href="' + decodeURIComponent(node.url) + '"]');
    if (activeNode) {
      activeNode.parentNode.classList.add('is-active');
    }
  }

  //
  // State Management
  //

  // Runs the search with a given query
  function handleSearchWithQuery(q) {
    searchField.value = q;

    // Every search action can be navigated to, just the "q" part changes.
    history.replaceState(
      {node: history.current.node, search: q},
      '',
      window.location.pathname + '?q=' + q
    );

    tree.search(q).then((results) => {
      return tree.filteredBy(results).root;
    }).then((root) => {
      nav.innerHTML = renderNav(root, window.location.pathname);
    });
  }

  searchField.addEventListener('input', function() {
    handleSearchWithQuery(this.value);
  });
  searchClear.addEventListener('click', function() {
    handleSearchWithQuery("");
  });

  // Loads the node when a link the nav is clicked
  // and updates session history (url)
  let handleNav = function(ev) {
    // If CMD is pressed default behavior is triggered (open link in new tab)
    if (event.metaKey) {
      return;
    }

    ev.preventDefault();
    loadNodeWithPath(this.pathname, true);
  };

  // Restore previouse state.
  window.onpopstate = function(ev) {
    if (ev.state) {
      loadNodeWithPath(event.state.path, false);
      handleSearchWithQuery(event.state.search);
    }
  };

  // Foucs the search field when pressing CMD + k.
  window.addEventListener('keydown', function(ev) {
    if (ev.key === 'k' && ev.metaKey) {
      ev.preventDefault();
      searchField.focus();
    }
  });


  // Calls the search when a link in text is clicked
  function handleTextLinkClick(ev) {
    // When the link starts with "search:" it is not a link to be followed, but
    // a query to be entered into the search bar.
    if (this.href.split(":")[0] === "search") {
      ev.preventDefault();
      handleSearchWithQuery(this.href.substring(7));
      return;
    }

    // If CMD is pressed default behavior is triggered (open link in new tab)
    if (event.metaKey) {
      return;
    }

    // Only handle local links
    if (this.host === window.location.host) {
      // Only handle the link, when it is not a file (== the last part of the path doesn’t contain a ".")
      if (this.href.split("/").pop().split(".").length === 1) {
        ev.preventDefault();
        loadNodeWithPath(this.pathname, true);
      }
    }
  }


});
