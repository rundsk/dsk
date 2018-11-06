/*!
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import Client from './dsk/Client.js';

export default class Search {
  constructor(field, clear, stats, tree, props) {
    this.field = field;
    this.clear = clear;
    this.stats = stats;
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

  setStats(total, took) {
    this.total = total;
    this.took = took;
  }

  render() {
    this.field.value = this.query;

    if (!this.query) {
      this.stats.innerHTML = '&nbsp;';
      this.clear.classList.add('hide');
    } else {
      this.stats.innerHTML = `${this.total} result${this.total !== 1 ? 's' : ''} in ${this.took / 1000}µs`;
      this.clear.classList.remove('hide');
    }
  }

  perform() {
    if (!this.query) {
      this.onFilter(this.tree.root, this.query);
      this.render();
    } else {
      Client.filter(this.query)
        .then((res) => {
          let urls = res.nodes.map(n => n.url);

          this.onFilter(this.tree.filteredBy(urls).root, this.query);
          this.total = res.total;
          this.took = res.took;
          this.render();
        });
    }
  }
}
