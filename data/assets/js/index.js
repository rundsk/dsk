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
  let search = $1('.search-field');

  let tree = {};
  let list = {};

  function flattenTree(node) {
    let results = [];

    if (!node.children) {
      return results;
    }
    for (let child of node.children) {
      results.push(child);
      results = results.concat(flattenTree(child));
    }
    return results;
  }

  // Gets the tree and creates the nav structure
  fetch('/api/tree').then((res) => {
    return res.json();
  }).then((json) => {
    // Initialize our primary data structure. Convert tree to list once, so we
    // don't have to do it each time search is called. Tree is our single source
    // of truth.
    tree = json.data;
    list = flattenTree(tree.root);

    renderNav(tree);
  });

  // Loads the node based on url
  let handleUrl = function(url) {
    fetch(url).then((res) => {
      return res.text();
    }).then((html) => {
      $1('main').innerHTML = html;
      handleKeywords();
    });
  };

  // Initial check for route and load node
  if (window.location.pathname !== '/') {
    let url = window.location.protocol + '//' +
      window.location.host + '/tree' +
      window.location.pathname;
    handleUrl(url);
  }

  // Runs the search from the input field
  let handleSearch = function(ev) {
    if (this.value !== "") {
      runSearch(list, this.value);
    } else {
      renderNav(tree);
    }
  };

  // Runs the search with a given query
  let handleSearchWithQuery = function(q) {
    search.value = q;
    if (q !== "") {
      runSearch(list, q);
    } else {
      renderNav(tree);
    }
  };

  // Clears the search field
  let clearSearch = function() {
    handleSearchWithQuery("");
  };

  search.addEventListener("input", handleSearch);
  $1('.search-clear').addEventListener("click", clearSearch);

  // Loads the node when a link the nav is clicked
  // and updates session history (uri)
  let handleNav = function(ev) {
    ev.preventDefault();
    fetch(this.href).then((res) => {
      return res.text();
    }).then((html) => {
      let uri = this.href.split('tree').pop();
      history.pushState(null, '', uri);
      $1('main').innerHTML = html;
      handleKeywords();
    });
  };

  // Calls the search when a keyword is clicked
  let handleKeywordClick = function(ev) {
    handleSearchWithQuery(ev.target.innerHTML);
  };

  // Attaches a click-Event to every keyword
  let handleKeywords = function() {
    for (let k of $('.keyword')) {
      k.addEventListener("click", handleKeywordClick);
    }
  };

  // Runs the search and rebuilds the nav
  let runSearch = function(list, query) {
    let options = {
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
    };

    let fuse = new Fuse(list, options);
    let result = fuse.search(query);

    renderNav(tree, result);
  };

  // Renders the nav structure
  let renderNav = function(tree, searchResult) {
    nav.innerHTML = '';

    tree.root.keep = checkIfNodeShouldBeKept(tree.root, searchResult);

    let list = createList(tree.root);
    let ul = document.createElement('ul');

    // Append list withouth root node (a bit hacky)
    if (list) {
      nav.appendChild(list.childNodes[1]);
    }
  };

  // If a searchResult is given, checks for each node if it exists in
  // the searchResult and should therefore be kept.
  let checkIfNodeShouldBeKept = function(node, filterBy) {
    var keep = false;

    if (filterBy !== undefined) {
      if (node.children !== null) {

        // Iterate over children, if one of the children should be kept, this node should be kept
        for (var child in node.children) {
            let keepChild = checkIfNodeShouldBeKept(node.children[child], filterBy);
            if (keepChild) {
              keep = true;
              node.keep = true;
            }
        }

        // If this parent node itself is in the searchResults, it should be kept (with all its children)
        if (filterBy && node.url !== "/") {
          for (let i of filterBy) {
            if (i.url == node.url) {
              keep = true;

              for (let child in node.children) {
                  checkIfNodeShouldBeKept(node.children[child], undefined);
              }
            }
          }
        }

        node.keep = keep;
        return keep;
      } else {
        // If this leaf node itself is in the searchResults, it should be kept
        if (filterBy && node.url !== "/") {
          for (let i of filterBy) {
            if (i.url == node.url) {
              keep = true;
            }
          }

          if (keep === true) {
            node.keep = true;
            return true;
          } else {
            node.keep = false;
            return false;
          }
        }

      }
    } else {
      // When no searchResult is given, all nodes should be kept.
      if (node.children !== null) {
        // Make sure all children are kept
        for (let child in node.children) {
            checkIfNodeShouldBeKept(node.children[child], undefined);
        }
      }

      node.keep = true;
      return true;
    }
  };

  // Turns the given data into a "ul li" structure
  let createList = function(node) {
    if (node.keep !== false) {
      if (node.children !== null) {
        let li = document.createElement('li');
        let a  = document.createElement('a');

        a.href = '/tree/' + node.url;
        a.innerHTML = node.title;
        a.addEventListener('click', handleNav);
        li.appendChild(a);

        let ul = document.createElement('ul');
        li.appendChild(ul);

        for (var child in node.children) {
            childList = createList(node.children[child]);
            if (childList) {
              ul.appendChild(childList);
            }
        }

        return li;
      } else {
        let li = document.createElement('li');
        let a  = document.createElement('a');

        a.href = '/tree/' + node.url;
        a.innerHTML = node.title;
        a.addEventListener('click', handleNav);
        li.appendChild(a);

        return li;
      }
    }
  };
});
