/*!
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Client: false */

class Search {
  constructor(field, clear, tree, props) {
    this.field = field;
    this.clear = clear;
    this.tree = tree;

    this.query = field.value || '';

    this.onFilter = props.onFilter;

    let _this = this;
    this.field.addEventListener('input', function() {
      _this.setQuery(this.value);
      // Skip render, as the value is just read from there.
      _this.perform();
    });

    this.clear.addEventListener('click', () => {
      this.setQuery('');
      this.render();
      this.perform();
    });

    // Focus the search field when pressing CMD + k.
    window.addEventListener('keydown', (ev) => {
      if (ev.key !== 'k' || !ev.metaKey) {
        return;
      }
      ev.preventDefault();
      this.field.focus();
    });
  }

  setOnFilter(onFilter) {
    this.onFilter = onFilter;
  }

  setQuery(q) {
    this.query = q;
  }

  render() {
    this.field.value = this.query;
  }

  perform() {
    if (!this.query) {
      this.onFilter(this.tree.root, this.query);
    } else {
      Client.search(this.query).then((results) => {
        this.onFilter(this.tree.filteredBy(results).root, this.query);
      });
    }
  }
}
