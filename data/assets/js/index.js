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

  var data= {};

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
    data = json.data.nodeList;
    renderNav(data[0]);

  });

  let renderNav = function(data) {
    nav.innerHTML = '';
    let list = createList(data);
    let ul = document.createElement('ul');

    // Append full list
    //ul.appendChild(list);
    //nav.appendChild(ul);

    // Append list withouth root node (a bit hacky)
    list.querySelector("li a").remove();
    nav.appendChild(list.childNodes[0]);
  }

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
