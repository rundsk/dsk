/**
 * Copyright 2017 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

// Client for accessing the dsk APIv1.
class Client {
  static hello() {
    return this.fetch('/api/v1/hello');
  }

  static tree() {
    return this.fetch('/api/v1/tree');
  }

  // Returns node for given relative URL path.
  static get(url) {
    if (url.charAt(0) === '/') {
      url = url.substring(1);
    }
    if (url.charAt(url.length - 1) === '/') {
      url = url.slice(0, -1);
    }
    return this.fetch(`/api/v1/tree/${url}`);
  }

  // Performs a search against the tree and returns the URLs
  // from the nodes included in the result set. Use filteredBy()
  // to create a new tree view.
  static search(q) {
    return this.fetch(`/api/v1/search?q=${encodeURIComponent(q)}`);
  }

  // Wraps fetch API, to fail promise when there is a network issue (catch)
  // as well as when we a HTTP response status indicating an error. Unwrapped
  // promises to make it clear, what happens when.
  //
  // TODO: Support Basic Auth
  //       https://github.com/facebook/react-native/issues/3321
  //       https://stackoverflow.com/questions/30203044/using-an-authorization-header-with-fetch-in-react-native
  static fetch(url) {
    return new Promise((resolve, reject) => {
      fetch(url)
        .then((res) => {
          if (res.ok) {
            res.json().then(json => resolve(json.data));
          } else {
            reject(new Error(`API request for '${url}' failed :-S: ${res.statusText}`));
          }
        })
        .catch(err => reject(err));
    });
  }
}
