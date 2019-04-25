/*!
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Prism: false */

import Client from './dsk/Client.js';

export default class Page {
  constructor(main, props) {
    this.main = main;

    this.onSearch = props.onSearch;
    this.onNavigate = props.onNavigate;
    this.handleClicks = this.handleClicks.bind(this);

    main.addEventListener('click', this.handleClicks);
  }

  setBaseTitle(title) {
    this.baseTitle = title;
  }

  setNode(n) {
    this.node = n;
  }

  handleClicks(ev) {
    if (ev.metaKey || !ev.target || ev.target.nodeName !== 'A') {
      return;
    }
    let t = ev.target;
    let p = ev.target.parentNode;

    // Tag-links are not links to be followed, but queries
    // to be entered into the search bar. They usually have
    // the form:
    // /tree/Blocks/Entering-Data/RadioButtonGroup?q=Control
    if (p.classList.contains('tags__tag')) {
      ev.preventDefault();
      let url = new URL(t.href);
      let q = url.searchParams.get('q');

      this.onSearch(q);
      return;
    }
    if (p.classList.contains('crumbs__crumb')) {
      ev.preventDefault();
      let url = new URL(t.href);

      this.onNavigate(url.pathname.replace('/tree/', ''));
      return;
    }

    // Differentiate between node links and other links.
    if (t.dataset.node) {
      ev.preventDefault();
      this.onNavigate(t.dataset.node);
      return;
    }

    if (p.parentNode.classList.contains('children-table__child')) {
      ev.preventDefault();
      let url = new URL(t.href);

      this.onNavigate(url.pathname.replace('/tree/', ''));
      return;
    }
  }

  render() {
    if (!this.node.url) {
      document.title = this.baseTitle;
    } else {
      document.title = `${this.baseTitle}: ${this.node.title}`;
    }

    this.main.innerHTML = '';

    // Title Container
    let titleContainer = document.createElement('section');
    titleContainer.classList.add('title-container');
    this.main.appendChild(titleContainer);

    // Description Container
    let descriptionContainer = document.createElement('section');
    descriptionContainer.classList.add('description-container');
    this.main.appendChild(descriptionContainer);

    // Tab Container
    let tabContainer = document.createElement('section');
    tabContainer.classList.add('tab-container');
    this.main.appendChild(tabContainer);

    // Always display the tab
    let switches = document.createElement('section');
    switches.classList.add('doc-switches');
    tabContainer.appendChild(switches);

    // Metadata Container
    let metaDataContainer = document.createElement('section');
    metaDataContainer.classList.add('meta-data-container');
    this.main.appendChild(metaDataContainer);

    // Crumbs
    let crumbs = document.createElement('ul');
    crumbs.classList.add('crumbs');
    crumbs.classList.add('t-gamma-sans');
    titleContainer.appendChild(crumbs);

    this.node.crumbs.forEach((crumb) => {
      let li = document.createElement('li');
      li.classList.add('crumbs__crumb');
      crumbs.appendChild(li);

      let a = document.createElement('a');
      a.href = `/tree/${crumb.url}`;
      a.innerText = crumb.title;
      li.appendChild(a);
    });

    // Title
    let title = document.createElement('h1');
    title.classList.add('title');
    title.classList.add('t-alpha-sans-bold');
    title.innerText = this.node.title;
    titleContainer.appendChild(title);

    // Description
    if (this.node.description) {
      let description = document.createElement('section');
      description.classList.add('description');
      description.classList.add('t-beta-sans');
      description.innerHTML = this.node.description;
      descriptionContainer.appendChild(description);
    }

    // Tags
    if (this.node.tags.length) {
      let tags = document.createElement('ul');
      tags.classList.add('tags');
      tags.classList.add('t-delta-sans-bold');

      this.node.tags.forEach((tag) => {
        let li = document.createElement('li');
        li.classList.add('tags__tag');

        let a = document.createElement('a');
        a.href = `/tree/${this.node.url}?q=${tag}`;
        a.innerText = tag;

        li.appendChild(a);
        tags.appendChild(li);
      });
      descriptionContainer.appendChild(tags);
    }

    // Info Area
    let info = document.createElement('section');
    info.classList.add('info');
    this.main.appendChild(info);

    // Docs Area
    let docs = document.createElement('section');
    docs.classList.add('docs');
    info.appendChild(docs);

    // Doc Switcher
    if (this.node.docs.length > 1) {
      this.node.docs.sort((a, b) => {
        if (a.title.toLowerCase() === 'readme') {
          return -1;
        }
        if (b.title.toLowerCase() === 'readme') {
          return 1;
        }
        return 0;
      });

      this.node.docs.forEach((doc, index) => {
        let a = document.createElement('a');
        a.classList.add('doc-switch');
        if (index === 0) {
          a.classList.add('active');
        }
        a.href = `#doc-${index}`;
        a.classList.add('t-gamma-sans');
        a.innerText = doc.title;

        a.addEventListener('click', (ev) => {
          ev.preventDefault();

          this.main.querySelector('.doc-switch.active').classList.remove('active');
          a.classList.add('active');

          this.main.querySelectorAll('.doc').forEach((el) => {
            el.classList.toggle('hide', el.id !== `doc-${index}`);
          });
        });
        switches.appendChild(a);
      });
    }

    // Docs
    if (this.node.docs.length) {
      this.node.docs.forEach((doc, index) => {
        let text = document.createElement('article');
        text.classList.add('doc');
        text.innerHTML = doc.html;

        if (index !== 0) {
          text.classList.add('hide');
        }
        text.id = `doc-${index}`;
        docs.appendChild(text);
      });
    } else {
      // With no docs, Contents tab option is showed
      let a = document.createElement('a');

      a.innerText = 'Contents';

      a.classList.add('doc-switch');
      a.classList.add('active');
      a.classList.add('t-gamma-sans');

      a.href = '#table';
      switches.appendChild(a);

      a.addEventListener('click', (ev) => {
        ev.preventDefault();

        this.main.querySelector('.doc-switch.active').classList.remove('active');

        a.classList.add('active');

        this.main.querySelector('.children-table').classList.toggle('hide');
        this.main.querySelectorAll('.doc').forEach((el) => {
          el.classList.add('hide');
        });
      });

      let dir = document.createElement('table');
      dir.classList.add('children-table');
      info.appendChild(dir);

      this.node.children.forEach((c) => {
        let row = document.createElement('tr');
        row.classList.add('children-table__child');
        dir.appendChild(row);

        let cell0 = document.createElement('td');
        row.appendChild(cell0);

        let at = document.createElement('a');
        at.innerText = c.title;
        at.href = `/tree/${c.url}`;
        cell0.appendChild(at);

        let cell1 = document.createElement('td');
        row.appendChild(cell1);

        let ad = document.createElement('a');
        ad.href = `/tree/${c.url}`;
        cell1.appendChild(ad);

        Client.get(c.url).then((item) => {
          ad.innerText = item.description || '';
        });
      });
    }

    // Meta data
    if (this.node.authors.length !== 0) {
      let h1 = document.createElement('h1');
      h1.classList.add('meta-data-container__title');
      h1.classList.add('t-delta-sans-bold');
      h1.innerText = 'author';

      if (this.node.authors.length > 1) {
        h1.innerText += 's';
      }
      metaDataContainer.appendChild(h1);

      this.node.authors.forEach((author) => {
        let p = document.createElement('p');
        p.classList.add('meta-data-container__info');
        p.classList.add('t-gamma-sans');
        p.innerText = author.name;
        metaDataContainer.appendChild(p);
      });
    }

    if (this.node.version) {
      let h1 = document.createElement('h1');
      h1.classList.add('meta-data-container__title');
      h1.classList.add('t-delta-sans-bold');
      h1.innerText = 'version';
      metaDataContainer.appendChild(h1);

      let p = document.createElement('p');
      p.classList.add('meta-data-container__info');
      p.classList.add('t-gamma-sans');
      p.innerText = this.node.version;
      metaDataContainer.appendChild(p);
    }

    if (this.node.modified) {
      let h1 = document.createElement('h1');
      h1.classList.add('meta-data-container__title');
      h1.classList.add('t-delta-sans-bold');
      h1.innerText = 'last changed';
      metaDataContainer.appendChild(h1);

      let p = document.createElement('p');
      p.classList.add('meta-data-container__info');
      p.classList.add('t-gamma-sans');

      let modified = (new Date(this.node.modified * 1000)).toLocaleDateString();
      p.innerText = modified;
      metaDataContainer.appendChild(p);
    }

    // Downloads
    if (this.node.downloads.length) {
      let downloads = document.createElement('aside');
      downloads.classList.add('downloads');
      this.main.appendChild(downloads);

      let h1 = document.createElement('h1');
      h1.classList.add('downloads__headline');
      h1.classList.add('t-delta-sans-bold');
      h1.innerText = 'assets';
      downloads.appendChild(h1);

      let downloadsList = document.createElement('ul');
      downloads.appendChild(downloadsList);

      this.node.downloads.forEach((c) => {
        let li = document.createElement('li');
        li.classList.add('downloads__item');
        li.classList.add('t-gamma-sans-bold');
        downloadsList.appendChild(li);

        let a = document.createElement('a');
        a.href = `/api/v1/tree/${c.url}`;
        li.appendChild(a);

        let d = document.createElement('div');
        d.innerText = c.name;
        d.classList.add('downloads__item-title');
        a.appendChild(d);

        let modified = (new Date(c.modified * 1000)).toLocaleDateString();

        let p = document.createElement('p');
        p.classList.add('downloads__item-info');
        p.classList.add('t-gamma-sans');
        p.innerText = `${(c.size / 102400).toFixed(2)} MB — ${modified}`;
        a.appendChild(p);
      });
    }

    // Show Source
    let a = document.createElement('a');
    a.classList.add('source__item');
    a.classList.add('doc-switch');
    a.classList.add('doc-switch--source');
    a.innerText = 'Source';
    a.href = '#source-code';
    switches.appendChild(a);

    let source = document.createElement('div');
    let wrap = document.createElement('article');
    wrap.classList.add('hide');
    wrap.classList.add('doc');
    wrap.id = 'source-code';
    source.appendChild(wrap);

    let pre = document.createElement('pre');
    pre.classList.add('source__code');
    wrap.appendChild(pre);

    let code = document.createElement('code');
    code.classList.add('language-json');

    code.innerHTML = Prism.highlight(
      JSON.stringify(this.node, null, 2),
      Prism.languages.javascript,
      'javascript',
    );
    pre.appendChild(code);

    let more = document.createElement('a');
    wrap.classList.add('source__diy');
    more.href = 'https://github.com/atelierdisko/dsk/tree/master#building-your-own-frontend';
    more.target = 'new';
    more.innerText = 'Build your own frontend';
    wrap.appendChild(more);
    docs.appendChild(wrap);

    a.addEventListener('click', (ev) => {
      ev.preventDefault();

      let activeDoc = this.main.querySelector('.doc-switch.active');
      if (activeDoc) {
        this.main.querySelector('.doc-switch.active').classList.remove('active');
      }
      a.classList.add('active');

      if (this.node.docs.length) {
        this.main.querySelectorAll('.doc').forEach((el) => {
          el.classList.add('hide', el.id !== 'source-code');
        });
      } else {
        this.main.querySelector('.children-table').classList.add('hide');
      }

      wrap.classList.toggle('hide');
    });
  }
}
