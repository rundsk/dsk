/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Nav: false */
/* globals Page: false */
/* globals Search: false */
/* globals Tree: false */
/* globals Client: false */
/* globals URLSearchParams: false */

document.addEventListener('DOMContentLoaded', () => {
  const $1 = document.querySelector.bind(document);

  let tree = new Tree();

  let nav = new Nav($1('.tree-nav'), {
    onNavigate: navigateToNode,
  });
  let search = new Search($1('.search__field'), $1('.search__clear'), $1('.search__stats'), tree, {
    // onFilter set later below.
  });
  let page = new Page($1('main'), {
    onSearch: (q) => {
      search.setQuery(q);
      search.render();
      search.perform();
    },
    onNavigate: navigateToNode,
  });

  // Capture page and nav.
  search.setOnFilter((root, q) => {
    history.replaceState(
      {
        node: page.node,
        search: q || '',
      },
      '',
      // Every search action can be navigated to, just the "q" part changes.
      q ? `${window.location.pathname}?q=${q}` : window.location.pathname,
    );

    nav.setRoot(root);
    nav.render();
  });

  // Initializes the tree, the left hand tree navigation and displays
  // the currently route selected node - if any.
  (function initialState() {
    let params = new URLSearchParams(window.location.search);

    // Fetch in parallel.
    let hello = Client.hello();
    let get = Client.get(window.location.pathname.replace('/tree/', ''));
    let sync = tree.sync();

    hello.then((data) => {
      // Representing.
      console.log('%c DSK ', 'background-color: #0E26FC; color: white;');
      console.log(`Version ${data.version}`);

      $1('.project-name').innerText = data.project;
      page.setBaseTitle(`${document.title} ${data.project}`);
      // Will render page in restoreState().

      Promise.all([get, sync]).then((vals) => {
        restoreState(params.get('q'), vals[0]);
      });
    });
  }());

  // Restore previous state, when navigating back and forth.
  window.addEventListener('popstate', (ev) => {
    if (ev.state) {
      restoreState(ev.state.search, ev.state.node);
    }
  });

  function restoreState(q, n) {
    search.setQuery(q);
    page.setNode(n);
    nav.setActiveNode(n);

    search.render();
    page.render();
    // Must wait until we have nav, to activateNode.

    if (!q) {
      nav.setRoot(tree.root);
      nav.render();
    } else {
      Client.search(q).then((results) => {
        nav.setRoot(tree.filteredBy(results.urls).root);
        nav.render();
        search.setStats(results.total, results.took);
        search.render();
      });
    }
  }

  // Load the node based on node URL path.
  function navigateToNode(nurl, replaceState) {
    return Client.get(nurl).then((n) => {
      let params = new URLSearchParams(window.location.search);
      let q = params.get('q');

      history[replaceState ? 'replaceState' : 'pushState'](
        { node: n, search: q },
        '',
        q ? `/tree/${nurl}?q=${q}` : `/tree/${nurl}`,
      );

      page.setNode(n);
      nav.setActiveNode(n);

      page.render();
      nav.render();
    });
  }
});
