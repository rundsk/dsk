/**
 * Copyright 2017 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

// Client for accessing the DSK APIv2.
export default class Client {
  static hello() {
    return this.fetch('/api/v2/hello');
  }

  // Returns a WebSocket connection to the messages endpoint. Asummes it
  // is reachable over TLS, when we have been loaded using it.
  static messages() {
    let protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    let host = window.location.hostname;
    let port = (window.location.port ? `:${window.location.port}` : '');

    return new WebSocket(`${protocol}://${host}${port}/api/v2/messages`);
  }

  static tree() {
    return this.fetch('/api/v2/tree');
  }

  // Returns node for given relative URL path. Will automatically strip leading
  // and trailing slashes from the given node URL to turn it into a valid node
  // URL for lookup.
  static get(url) {
    if (url.charAt(0) === '/') {
      url = url.substring(1);
    }
    if (url.charAt(url.length - 1) === '/') {
      url = url.slice(0, -1);
    }
    return this.fetch(`/api/v2/tree/${url}`);
  }

  // Performs a full text search against the tree and returns the response
  // unfiltered.
  static search(q) {
    let params = new URLSearchParams();

    params.set('q', encodeURIComponent(q));

    return this.fetch(`/api/v2/search?${params.toString()}`);
  }

  // Performs a narrow search against the tree. The returned nodes together
  // with `Tree.filteredBy()` can be used to create a new filtered tree view:
  //
  // ```
  // Client.filter('foo')
  //   .then((res) => {
  //      return res.nodes.map(n => n.url);
  //   })
  //   .then((urls) => {
  //      return Tree.filteredBy(urls);
  //   });
  // ```
  static filter(q, useWideIndex = false) {
    let params = new URLSearchParams();

    params.set('q', encodeURIComponent(q));
    if (useWideIndex) {
      params.set('index', 'wide');
    }

    return this.fetch(`/api/v2/filter?${params.toString()}`);
  }

  // Performs API requests. Fail promise when there is a network issue (catch)
  // as well as when we a HTTP response status indicating an error. Using plain
  // XHR for better browser support and easier basic auth handling.
  static fetch(url) {
    return new Promise((resolve, reject) => {
      let xhr = new XMLHttpRequest();

      xhr.addEventListener('readystatechange', () => {
        if (xhr.readyState === 4) {
          let first = xhr.status.toString().charAt(0);
          if (first !== '2' && first !== '3') {
            try {
              reject(new Error(`API request for '${url}' failed :-S: ${JSON.parse(xhr.responseText).message}`));
            } catch (e) {
              reject(new Error(`API request for '${url}' failed :-S: ${xhr.statusText}`));
            }
            return;
          }
          try {
            resolve(JSON.parse(xhr.responseText));
          } catch (e) {
            reject(new Error(`API request for '${url}' succeeded, but failed to parse response :-S: ${e}`));
          }
        }
      });
      xhr.addEventListener('error', (ev) => {
        reject(new Error(`API request for '${url}' failed :-S: ${ev}`));
      });
      xhr.open('GET', url);
      xhr.setRequestHeader('Accept', 'application/json');
      xhr.send();
    });
  }
}
