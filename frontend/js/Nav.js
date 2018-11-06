/*!
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

export default class Nav {
  constructor(nav, props) {
    this.nav = nav;
    this.onNavigate = props.onNavigate;
  }

  setRoot(root) {
    this.root = root;
  }

  setActiveNode(node) {
    this.activeNode = node;
  }

  render() {
    this.nav.innerHTML = '';

    if (this.root.children.length) {
      this.nav.appendChild(
        this.renderList(this.root, this.activeNode).childNodes[1],
      );
    }
  }

  // Recursively turns the given tree data into a "ul li" structure.
  renderList(n, activeNode) {
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
      this.onNavigate(n.url);
    });

    li.appendChild(a);

    if (n.children.length) {
      li.appendChild(ul);
    }
    n.children.forEach((c) => {
      let cl = this.renderList(c, activeNode);
      if (cl) {
        ul.appendChild(cl);
      }
    });
    return li;
  }
}
