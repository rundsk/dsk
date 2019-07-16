/**
 * Copyright 2017 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import Client from './Client.js';

// Tree data structure for working with responses from the dsk API.
//
// https://en.wikipedia.org/wiki/Tree_traversal
export default class Tree {
  constructor(root) {
    this.root = root;
  }

  // One way sync: updates the data from backend source.
  sync() {
    return Client.tree().then(data => {
      this.root = data.root;
    });
  }

  // Returns a flat list of all nodes in the tree. The "node" parameter is for
  // rescursive invocation and should be null when called initally.
  flatten(node = null) {
    let list = [];

    ((node || this.root).children || []).each(child => {
      list.push(child);
      list = list.concat(this.flatten(child));
    });
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
  filteredBy(selectedURLs = []) {
    let tree = new Tree(JSON.parse(JSON.stringify(this.root))); // deep clone

    if (selectedURLs) {
      let check = n => selectedURLs.includes(n.url) || n.children.some(check);

      let select = n => {
        if (selectedURLs.includes(n.url)) {
          return true;
        }

        n.children = n.children.filter(select);
        return n.children.some(check);
      };
      tree.root.children = tree.root.children.filter(select);
    }

    return tree;
  }
}
