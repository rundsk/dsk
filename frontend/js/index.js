/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Tree: false */
/* globals Client: false */
/* globals Changes: false */
/* globals URLSearchParams: false */

document.addEventListener('DOMContentLoaded', function() {
  const $ = document.querySelectorAll.bind(document);
  const $1 = document.querySelector.bind(document);

  const baseTitle = document.title;

  let tree = new Tree();
  let changes = Client.changes();

  let activeNode = null;
  let activeTree = null;
  let activeSearch = null;

  let main = $1('main');
  let nav = $1('.tree-nav');
  let searchField = $1('.search__field');
  let searchClear = $1('.search__clear');

  //
  // Changes
  //

  changes.onmessage = (ev) => {
    // TODO: Is the current node displayed still present?

    tree.sync().then(() => {
      renderNav(tree.root, activeNode);
    });

    console.log(ev);
  };

  changes.onclose = (ev) => {
    console.log('Disconnected');
  };

  changes.onerror = (err) => {
    console.log('Error:', err);
  };


  //
  // Global State
  //

  // Initializes the tree, the left hand tree navigation and displays
  // the currently route selected node - if any.
  (function initialState() {
    let params = new URLSearchParams(window.location.search);
    let get = Client.get(window.location.pathname.replace('/tree/', ''));
    let sync = tree.sync();

    Promise.all([get, sync]).then((vals) => {
      restoreState(params.get('q'), vals[0]);
    });
  }());


  // Restore previous state, when navigating back and forth.
  window.addEventListener('popstate', (ev) => {
    if (ev.state) {
      restoreState(ev.state.search, ev.state.node);
    }
  });

  function restoreState(q, n) {
    searchField.value = q;

    setTitle(n);
    renderNode(n);
    // Must wait until we have nav, to activateNode.
    activeNode = n;

    if (!q) {
      renderNav(tree.root, activeNode);
    } else {
      Client.search(q).then((results) => {
        renderNav(tree.filteredBy(results).root, activeNode);
      });
    }
  }

  function setTitle(n) {
    document.title = (n ? `${baseTitle}: ${n.title}` : baseTitle);
  }

  //
  // Node Display
  //

  // main.querySelectorAll('.text a, .crumbs a, .keywords a, .children-table a').forEach((el) => {
  main.addEventListener('click', (ev) => {
    if (ev.metaKey || !ev.target || ev.target.nodeName !== 'A') {
      return;
    }
    ev.preventDefault();

    // Tag-links are not links to be followed, but queries
    // to be entered into the search bar.

    // Differntiate between node links and other links.
    console.log('Clicked link in main!', ev.target);
  });

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

      setTitle(n);
      renderNode(n);
      activeNode = n;

      let active;
      active = nav.querySelector('.is-active');
      if (active) {
        active.classList.remove('is-active');
      }
      active = nav.querySelector(`a[href="/tree/${n.url}"]`);
      if (active) {
        active.parentNode.classList.add('is-active');
      }
    });
  }

  function renderNode(n) {
    main.innerHTML = '';

    let title = document.createElement('section');
    title.classList.add('title');
    title.innerHTML = `<h1>${n.title}</h1>`;

    main.appendChild(title);

    if (n.description) {
      let description = document.createElement('section');
      description.classList.add('description');
      description.innerHTML = n.description;

      main.appendChild(description);
    }

    if (n.tags.length) {
      let tags = document.createElement('section');
      tags.classList.add('tags');

      let ul = document.createElement('ul');

      n.tags.forEach((tag) => {
        let li = document.createElement('li');
        li.classList.add('tags__tag');

        let a = document.createElement('a');
        a.href = `/tree/${n.url}?q=${tag}`;
        a.innerText = tag;

        li.appendChild(a);
        ul.appendChild(li);
      });

      tags.appendChild(ul);
      main.appendChild(tags);
    }

    let info = document.createElement('div');
    info.classList.add('info');

    main.appendChild(info);

    let twoColumns = document.createElement('div');
    twoColumns.classList.add('two-columns');

    info.appendChild(twoColumns);

    if (n.docs[0]) {
      let text = document.createElement('div');
      text.classList.add('text');
      text.innerHTML = n.docs[0].html;

      twoColumns.appendChild(text);
    }
  }

  //
  // Tree Navigation
  //

  function renderNav(root, activeNode) {
    nav.innerHTML = '';

    if (root.children.length) {
      nav.appendChild(renderList(root, activeNode).childNodes[1]);
    }
  }

  // Recursively turns the given tree data into a "ul li" structure.
  function renderList(n, activeNode) {
    let li = document.createElement('li');
    let ul = document.createElement('ul');
    let a = document.createElement('a');

    if (activeNode && n.url === activeNode.url) {
      li.classList.add('is-active');
    }

    a.href = `/tree/${n.url}`;
    a.innerHTML = n.title;
    a.addEventListener('click', (ev) => {
      if (ev.metaKey) {
        return;
      }
      ev.preventDefault();
      navigateToNode(n.url);
    });

    li.appendChild(a);

    if (n.children.length) {
      li.appendChild(ul);
    }
    n.children.forEach((c) => {
      let cl = renderList(c, activeNode);
      if (cl) {
        ul.appendChild(cl);
      }
    });
    return li;
  }

  //
  // Search
  //

  searchField.addEventListener('input', function() {
    let q = this.value;

    history.replaceState(
      { node: activeNode, search: q },
      '',
      // Every search action can be navigated to, just the "q" part changes.
      q ? `${window.location.pathname}?q=${q}` : window.location.pathname,
    );

    if (!q) {
      renderNav(tree.root, activeNode);
    } else {
      Client.search(q).then((results) => {
        renderNav(tree.filteredBy(results).root, activeNode);
      });
    }
  });

  searchClear.addEventListener('click', () => {
    searchField.value = '';
    history.replaceState({ node: activeNode, search: '' }, '', window.location.pathname);
    renderNav(tree.root, activeNode);
  });

  // Focus the search field when pressing CMD + k.
  window.addEventListener('keydown', (ev) => {
    if (ev.key !== 'k' || !ev.metaKey) {
      return;
    }
    ev.preventDefault();
    searchField.focus();
  });
});
