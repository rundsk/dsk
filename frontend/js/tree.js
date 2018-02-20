/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

"use strict";

// Client to the "tree" part of the DSK API and data structure for holding
// responses.
//
// https://en.wikipedia.org/wiki/Tree_traversal
class Tree {

  constructor(root) {
    this.root = root;
  }

  // One way sync: updates the data from backend source.
  sync() {
    return fetch('/api/v1/tree').then((res) => {
      return res.json();
    }).then((json) => {
      this.root = json.data.root;
    });
  }

  // Returns a flat list of all nodes in the tree. The "node" parameter is for
  // rescursive invocation and should be null when called initally.
  flatten(node = null) {
    let list = [];

    for (let child of (node || this.root).children || []) {
      list.push(child);
      list = list.concat(this.flatten(child));
    }
    return list;
  }

  // Returns a new non-sparse tree instance selecting only given
  // nodes, their parents and all their children.
  //
  // Filters out any not selected nodes. Descends into branches first,
  // then works its way back up the tree filtering out any nodes, that
  // are not selected. For selection conditions see check().
  //
  // Selecting a leaf node, selects all parents. But not the siblings.
  //
  //           a*
  //
  //           b*
  //
  //      c!   d   e
  //
  // Selecting a node, always selects all its children.
  //
  //           a*
  //
  //           b!
  //
  //      c*   d*   e*
  //
  filteredBy(selectedNodes = []) {
    let tree = JSON.parse(JSON.stringify(this)); // deep clone
    let selectedURLs = selectedNodes.map((n) => n.url);

    let check = function(n) {
      return selectedURLs.includes(n.url) || n.children.some(check);
    };

    let select = function(n) {
      if (selectedURLs.includes(n.url)) {
        return true;
      }

      n.children = n.children.filter(select);
      return n.children.some(check);
    };
    tree.root.children = tree.root.children.filter(select);

    return tree;
  }

  // Returns node for given relative URL path.
  get(url) {
    if (url.charAt(0) === "/") {
      url = url.substring(1);
    }
    if (url.charAt(url.length - 1) === "/") {
      url = url.slice(0, -1);
    }
    return fetch('/api/v1/tree/' + url).then((res) => {
      return res.json();
    }).then((json) => {
      return json.data;
    });
  }

  search(q) {
    return fetch('/api/v1/search?q=' + encodeURIComponent(q)).then((res) => {
      return res.json();
    }).then((json) => {
      return json.data;
    });
  }


}
