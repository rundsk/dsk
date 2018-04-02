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
  const icon = 'data:image/png;base64,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';

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

  $1('a.project-name').addEventListener('click', (ev) => {
    ev.preventDefault();
    navigateToNode('');
  });

  // Initializes the tree, the left hand tree navigation and displays
  // the currently route selected node - if any.
  (function initialState() {
    let params = new URLSearchParams(window.location.search);

    // Fetch in parallel.
    let hello = Client.hello();
    let get = Client.get(window.location.pathname.replace('/tree/', ''));
    let messages = Client.messages();
    let sync = tree.sync();

    messages.addEventListener('message', (ev) => {
      let m = JSON.parse(ev.data);

      if (m.type === 'tree-synced') {
        let repage = Client.get(page.node.url);
        let resync = tree.sync();

        resync.then(() => {
          notify('Re-synchronized', m.text);

          // Will re-render nav through onFilter.
          search.perform();

          // Change this once we can check for node
          // presence through API.
          repage
            .then((n) => {
              page.setNode(n);
              page.render();
            })
            .catch((e) => {
              // Node has gone away, show root.
              if (e.message.endsWith('No such node')) {
                navigateToNode('');
              }
            });
        });
      }
    });

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

  // Notifies the client, asking for permission if necessary.
  function notify(title, body) {
    if (Notification.permission === 'granted') {
      let n = new Notification(title, { body, icon });

      setTimeout(n.close.bind(n), 2500);
      return;
    }
    if (Notification.permission === 'default') {
      Notification.requestPermission().then(() => {
        notify(title, body);
      });
    }
  }
});
