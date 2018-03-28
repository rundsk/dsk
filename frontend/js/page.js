/*!
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/* globals Client: false */
/* globals Prism: false */

class Page {
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
    document.title = `${this.baseTitle}: ${this.node.title}`;

    this.main.innerHTML = '';

    // Crumbs
    let crumbs = document.createElement('ul');
    crumbs.classList.add('crumbs');
    this.main.appendChild(crumbs);

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
    title.innerText = this.node.title;
    this.main.appendChild(title);

    // Description
    if (this.node.description) {
      let description = document.createElement('section');
      description.classList.add('description');
      description.innerHTML = this.node.description;

      this.main.appendChild(description);
    }

    // Tags
    if (this.node.tags.length) {
      let tags = document.createElement('ul');
      tags.classList.add('tags');

      this.node.tags.forEach((tag) => {
        let li = document.createElement('li');
        li.classList.add('tags__tag');

        let a = document.createElement('a');
        a.href = `/tree/${this.node.url}?q=${tag}`;
        a.innerText = tag;

        li.appendChild(a);
        tags.appendChild(li);
      });
      this.main.appendChild(tags);
    }

    // Docs
    if (this.node.docs.length > 1) {
      // Zwitch between documents
      let switches = document.createElement('section');
      switches.classList.add('doc-switches');
      this.main.appendChild(switches);

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
        a.href = `#doc-${index}`;
        a.innerText = doc.title;

        a.addEventListener('click', (ev) => {
          ev.preventDefault();

          this.main.querySelectorAll('.doc').forEach((el) => {
            el.classList.toggle('hide', el.id !== `doc-${index}`);
          });
        });
        switches.appendChild(a);
      });
    }

    // Info Area
    let info = document.createElement('section');
    info.classList.add('info');
    this.main.appendChild(info);

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
        info.appendChild(text);
      });
    } else {
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
          ad.innerText = c.description || '';
        });
      });
    }

    // Downloads
    if (this.node.downloads.length) {
      let downloads = document.createElement('aside');
      downloads.classList.add('downloads');
      info.appendChild(downloads);

      let h1 = document.createElement('h1');
      h1.classList.add('downloads__title');
      h1.innerText = 'Downloads';
      downloads.appendChild(h1);

      let downloadsList = document.createElement('ul');
      downloads.appendChild(downloadsList);

      this.node.downloads.forEach((c) => {
        let li = document.createElement('li');
        li.classList.add('download');
        downloadsList.appendChild(li);

        let a = document.createElement('a');
        a.href = `/api/v1/tree/${c.url}`;
        a.innerText = c.name;
        a.setAttribute('download', '');
        li.appendChild(a);
      });
    }

    // Show Source
    let source = document.createElement('section');
    source.classList.add('source');
    this.main.appendChild(source);

    let a = document.createElement('a');
    a.innerText = 'Source';
    a.href = '#source-code';
    source.appendChild(a);

    let wrap = document.createElement('div');
    wrap.classList.add('hide');
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

    a.addEventListener('click', (ev) => {
      ev.preventDefault();
      wrap.classList.toggle('hide');
    });
  }
}

