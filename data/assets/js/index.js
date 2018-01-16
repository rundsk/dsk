/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

document.addEventListener('DOMContentLoaded', function() {
  let $ = document.querySelectorAll.bind(document);
  let $1 = document.querySelector.bind(document);

  let nav = $1('.tree-nav');

  let tree = new Tree();

  let fuse = new Fuse([], {
    tokenize: false,
    matchAllTokens: true,
    threshold: 0.1,
    location: 0,
    distance: 100,
    maxPatternLength: 32,
    minMatchCharLength: 1,
    keys: [
      "title",
      "url",
      "meta.keywords"
    ]
  });

  let pageTitle = document.title;

  // Gets the tree and creates the nav structure.
  // Get the query from the current window path (handleSearchWithQuery will render the Nav).
  tree.sync()
    .then(() => {
      fuse.setCollection(tree.flatten());

      // Initial check for route and load node
      loadNodeWithPath(window.location.pathname, false);
      handleSearchWithQuery(window.location.search.substring(1));
    });

  // Load the node based on path
  let loadNodeWithPath = function(path, pushToHistory) {
    if (path.charAt(path.length - 1) !== "/") {
      path += "/";
    }

    fetch("/api/tree" + path).then((res) => {
      return res.text();
    }).then((html) => {
      markNodeInNavAsActiveWithPath(path);

      let state = { path: path, search: window.location.search.substring(1) };
      if (pushToHistory) {
        history.pushState(state, '', path + window.location.search);
      } else {
        history.replaceState(state, '', path + window.location.search);
      }

      // Set document title to the name of the node
      var title = path.split("/");
      title = decodeURIComponent(title[title.length - 2]);
      titleArray = /^\d+[_,-]{1}(.*)/.exec(title)

      if (title !== "") {
        if (titleArray) {
          document.title = pageTitle + ": " + titleArray[1];
        } else {
          document.title = pageTitle + ": " + title;
        }
      } else {
        document.title = pageTitle;
      }

      $1('main').innerHTML = html;
      handleTextLinks();
    });
  };

  // Runs the search with a given query
  let handleSearchWithQuery = function(q) {
    if ($1('.search__field').value !== q) {
      $1('.search__field').value = q;
    }

    // Add query to the url
    let state = { path: window.location.pathname, search: q };
    let url = window.location.origin + window.location.pathname + "?" + q;
    history.replaceState(state, '', url);

    if (q !== "") {
      renderNav(tree, fuse.search(q));
    } else {
      renderNav(tree);
    }

    markNodeInNavAsActiveWithPath(window.location.pathname);
  };

  $1('.search__field').addEventListener("input", function() {
    handleSearchWithQuery(this.value);
  });
  $1('.search__clear').addEventListener("click", function() {
    handleSearchWithQuery("");
  });

  // Loads the node when a link the nav is clicked
  // and updates session history (url)
  let handleNav = function(ev) {
    // If CMD is pressed default behavior is triggered (open link in new tab)
    if (event.metaKey) {
      return
    }

    ev.preventDefault();
    loadNodeWithPath(this.pathname, true);
  };

  // Calls the search when a link in text is clicked
  let handleTextLinkClick = function(ev) {

    // When the link starts with "search:" it is not a link to be followed, but a query to be entered into the search bar
    if (this.href.split(":")[0] === "search") {
      ev.preventDefault();
      handleSearchWithQuery(this.href.substring(7));
      return;
    }

    // If CMD is pressed default behavior is triggered (open link in new tab)
    if (event.metaKey) {
      return
    }

    // Only handle local links
    if (this.host === window.location.host) {
      // Only handle the link, when it is not a file (== the last part of the path doesn’t contain a ".")
      if (this.href.split("/").pop().split(".").length === 1) {
        ev.preventDefault();
        loadNodeWithPath(this.pathname, true);
      }
    }
  };

  // Attaches a click-Event to every link in text
  let handleTextLinks = function() {
    for (let k of $('.text a')) {
      k.addEventListener("click", handleTextLinkClick);
    }

    for (let k of $('.crumbs a')) {
      k.addEventListener("click", handleTextLinkClick);
    }

    for (let k of $('.keywords a')) {
      k.addEventListener("click", handleTextLinkClick);
    }

    for (let k of $('.children-table a')) {
      k.addEventListener("click", handleTextLinkClick);
    }
  };

  // Renders the nav structure
  let renderNav = function(tree, searchResults) {
    nav.innerHTML = '';
    var list;

    if (searchResults) {
      list = createList(tree.filteredBy(searchResults).root);
    } else {
      // When none selected, all nodes should be kept, resets view.
      list = createList(tree.root);
    }
    if (list) {
      // Append list withouth root node.
      nav.appendChild(list.childNodes[1]);
    }
  };


  // Turns the given data into a "ul li" structure
  let createList = function(node) {
    if (node.keep === false) {
      return;
    }

    let li = document.createElement('li');
    let a  = document.createElement('a');
    a.href = "/" + node.url;
    a.innerHTML = node.title;
    a.addEventListener('click', handleNav);

    if (node.isGhost) {
      a.classList.add('ghosted');
    }

    li.appendChild(a);

    let ul = document.createElement('ul');
    li.appendChild(ul);

    for (var child in node.children) {
      var childList = createList(node.children[child]);
      if (childList) {
        ul.appendChild(childList);
      }
    }

    return li;
  };

  let markNodeInNavAsActiveWithPath = function(path) {
    for (let a of $('.tree-nav li')) {
      a.classList.remove("is-active");
    }

    let activeNode = $1(".tree-nav li a[href='" + decodeURIComponent(path) + "']");
    if (activeNode) {
      activeNode.parentNode.classList.add("is-active");
    }
  };

  window.onpopstate = function(event) {
    if (event.state) {
      loadNodeWithPath(event.state.path, false);
      handleSearchWithQuery(event.state.search);
    }
  };

  window.addEventListener("keydown", function (event) {
    if (event.key === "k" && event.metaKey) { // CMD + k
      event.preventDefault();
      $1('.search__field').focus();
    }
  });

  $1("header a").addEventListener("click", handleTextLinkClick);
});
