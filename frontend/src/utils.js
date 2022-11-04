/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { useLocation } from 'react-router-dom';

// Constructs a url based on the given props, while maintaining
// the params from the current URL
export const constructURL = ({ node, source, activeTab }) => {
  if (!node) {
    node = window.location.pathname;
  }

  if (!node.startsWith('/')) {
    node = `/${node}`;
  }

  let urlParams = new URLSearchParams(window.location.search);

  if (source) {
    urlParams.set('v', source);
  }
  if (source === null) {
    urlParams.delete('v');
  }

  if (activeTab) {
    urlParams.set('t', activeTab);
  }
  if (activeTab === null) {
    urlParams.delete('t');
  }

  urlParams.sort();
  const params = urlParams.toString();

  return `${node}${params ? `?${params}` : ''}`;
};

// A custom hook that builds on useLocation to parse
// the query string for you.
export function useURL() {
  const location = useLocation();
  const query = new URLSearchParams(location.search);

  return {
    nodeURL: location.pathname.substring(1),
    search: location.search,
    source: query.get('v'),
    activeTab: query.get('t'),
  };
}

// via https://gist.github.com/hagemann/382adfc57adbd5af078dc93feef01fe1
export function slugify(string) {
  const a = 'àáäâãåăæçèéëêǵḧìíïîḿńǹñòóöôœṕŕßśșțùúüûǘẃẍÿź·/_,:;';
  const b = 'aaaaaaaaceeeeghiiiimnnnoooooprssstuuuuuwxyz------';
  const p = new RegExp(a.split('').join('|'), 'g');
  return string
    .toString()
    .toLowerCase()
    .replace(/\s+/g, '-') // Replace spaces with -
    .replace(p, (c) => b.charAt(a.indexOf(c))) // Replace special characters
    .replace(/&/g, '-and-') // Replace & with ‘and’
    .replace(/[^\w-]+/g, '') // Remove all non-word characters
    .replace(/--+/g, '-') // Replace multiple - with single -
    .replace(/^-+/, '') // Trim - from start of text
    .replace(/-+$/, ''); // Trim - from end of text
}

// via https://stackoverflow.com/a/30810322/
export function copyTextToClipboard(text) {
  var textArea = document.createElement('textarea');

  textArea.style.position = 'fixed';
  textArea.style.top = -1000;
  textArea.style.left = -1000;

  textArea.value = text;

  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    document.execCommand('copy');
  } catch (err) {
    console.log('Oops, unable to copy');
  }

  document.body.removeChild(textArea);
}
