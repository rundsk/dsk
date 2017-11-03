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

  let handleNav = function(ev) {
      ev.preventDefault();
      fetch(this.href).then((res) => {
        return res.text();
      }).then((html) => {
        $1('main').innerHTML = html;
      });
  };

  fetch('/api/tree').then((res) => {
    return res.json();
  }).then((json) => {
    // let ul = document.createElement('ul');
    //
    //
    // // TODO: built tree, currently a flat list
    // for (let n of json.data.nodeList) {
    //   let li = document.createElement('li');
    //   let a  = document.createElement('a');
    //   a.href = '/tree/' + n.url;
    //   a.innerHTML = n.url;
    //   a.addEventListener('click', handleNav);
    //
    //   li.appendChild(a);
    //   ul.appendChild(li);
    // }
    nav.innerHTML = '';

    let list = createList(json.data.nodeList[0])
    let ul = document.createElement('ul');
    ul.appendChild(list);
    nav.appendChild(ul);
  });

  let createList = function(obj) {
    if (obj.children !== null) {
      let li = document.createElement('li');
      let a  = document.createElement('a');
      a.href = '/tree/' + obj.url;
      a.innerHTML = obj.title;
      a.addEventListener('click', handleNav);
      li.appendChild(a);

      let ul = document.createElement('ul');
      li.appendChild(ul);

      for (var child in obj.children) {
          ul.appendChild(createList(obj.children[child]));
      }
      return li;
    } else {
      let li = document.createElement('li');
      let a  = document.createElement('a');
      a.href = '/tree/' + obj.url;
      a.innerHTML = obj.title;
      a.addEventListener('click', handleNav);

      li.appendChild(a);
      return li;
    }
  }
});
