/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useState, useEffect, useRef } from 'react';
import { routeNode, BaseLink } from 'react-router5';
import { useGlobal } from 'reactn';
import { Helmet } from 'react-helmet';

import { Client } from '@atelierdisko/dsk';
import TreeNavigation from './TreeNavigation';

import './Variables.css';
import './App.css';
import Page from './Page';
import ErrorPage from './ErrorPage';
import Search from './Search';

import HamburgerIcon from './HamburgerIcon.svg';
import CloseIcon from './CloseIcon.svg';

// System.register([], function (exports) {
//   'use strict';
//   return {
//     execute: function () {
//       function Button(props) {
//         return React.createElement("button", null, "Hello DSK!");
//       }
//       exports('Button', Button);
//     }
//   };
// });
class DSKBundleRegistry {
  constructor() {
    this.modules = [];
    this.exports = {};
  }

  register(unused, fn) {
    this.modules.push(fn);
  }

  load(name) {
    if (this.exports[name]) {
      return this.exports[name];
    }
    let exportfn = (name, fn) => {
      this.exports[name] = fn;
    }
    this.modules.forEach((m) => {
      let bundle = m(exportfn);
      bundle.execute();
    });
    if (this.exports[name]) {
      return this.exports[name];
    }
  }
}

window.System = new DSKBundleRegistry();
window.React = React;

function App(props) {
  const [tree, setTree] = useState(null);
  const [node, setNode] = useState(null);
  const [error, setError] = useState(null);
  const [source, setSource] = useGlobal('source');
  const [availableSources, setAvailableSources] = useState([]);
  const socket = useRef();
  const onMessage = useRef();
  const [config, setConfig] = useGlobal('config');
  const [mobileSidebarIsActive, setMobileSidebarIsActive] = useState(false);

  // Establish a WebSocket connection and register a handler, that will trigger
  // a full re-render of the App, once we receive a sync message. We're
  // intentionally not displaying notifications, as we consider them to be too
  // intrusive.
  useEffect(() => {
    onMessage.current = ev => {
      let m = JSON.parse(ev.data);

      if (m.topic === `${source}.tree.synced`) {
        loadTree();

        // The node might have gone away.
        checkNode(source).then(isExistent => {
          if (isExistent) {
            loadNode();
          } else {
            console.log('Current node has gone away after tree has synced.');
            props.router.navigate('home', { ...props.route.params, v: source });
          }
        });
      }

      if (m.topic.includes('source.status.changed')) {
        Client.sources().then(data => {
          setAvailableSources(data.sources);
        });
      }
    };
  }, [props.route, source]);

  useEffect(() => {
    if (socket.current) {
      // Ensure we don't open multiple sockets.
      return;
    }
    console.log('Establishing WebSocket connection...');
    socket.current = Client.messages();
    socket.current.addEventListener('message', ev => {
      onMessage.current(ev);
    });
  }, [socket, onMessage]);

  useEffect(() => {
    Client.sources().then(data => {
      setAvailableSources(data.sources);

      var sourceToLoad = null;
      let sourceFromURL = props.route.params.v;

      // First we check if the source from the url exists
      if (sourceFromURL) {
        data.sources.forEach(v => {
          if (v.name === sourceFromURL) {
            sourceToLoad = sourceFromURL;
          }
        });
      }

      // Then we check if a live source exists
      if (!sourceToLoad) {
        data.sources.forEach(v => {
          if (v.name === 'live') {
            sourceToLoad = 'live';
          }
        });
      }

      // If not we take the first source we can find
      if (!sourceToLoad) {
        sourceToLoad = data.sources[0].name;
      }

      setSource(sourceToLoad);
    });
  }, []);

  function loadTree() {
    if (!source) {
      return;
    }
    Client.tree(source)
      .then(data => {
        setTree(data.root);
      })
      .catch(err => {
        console.log(`Failed to load tree: ${err}`);
      });
  }

  function loadNode() {
    if (!source) {
      return;
    }
    Client.get(nodeURLFromRouter(props.route), source)
      .then(data => {
        setNode({ ...data, source: source });
        setError(null);
      })
      .catch(err => {
        console.log(`Failed to set node data: ${err}`);
        setError('Design aspect not found.');
      });
  }

  function checkNode(source) {
    return Client.has(nodeURLFromRouter(props.route), source);
  }

  function nodeURLFromRouter(route) {
    switch (route.name) {
      case 'home':
        return ''; // Is a valid node URL.
      case 'node':
        return route.params.node;
      default:
        return null;
    }
  }

  function changeSource(newSource) {
    setSource(newSource);

    // Update URL
    props.router.navigate(props.route.name, { ...props.route.params, v: newSource }, { replace: true });

    // The node might have gone away.
    checkNode(newSource).then(isExistent => {
      if (!isExistent) {
        console.log('Current node has gone away after tree has synced.');
        props.router.navigate('home', { ...props.route.params, v: newSource });
      }
    });
  }

  // This hook may run several times. We might receive an empty configuration
  // object from the API. We must differentiate between this case and initially
  // empty object.
  useEffect(() => {
    if (config._populated) {
      return;
    }
    Client.config().then(data => {
      setConfig({
        ...data,
        _populated: true,
      });
    });
  }, [config, setConfig]);

  // Initialize tree navigation.
  useEffect(loadTree, [source]);

  // Load the current node being displayed. Reload it whenever the route changes.
  useEffect(loadNode, [props.route, source]);

  let content;
  if (error) {
    content = <ErrorPage>{error}</ErrorPage>;
  } else if (node) {
    content = (
      <Page {...node} activeTab={props.route.params.t || undefined} baseTitle={config.org + ' / ' + config.project} />
    );
  }

  let refToMain = React.createRef();

  return (
    <div className="app">
      <Helmet htmlAttributes={{ lang: config.lang }} />

      <button
        className="app_skip-to-content"
        onClick={() => {
          if (refToMain.current) {
            refToMain.current.focus();
          }
        }}
      >
        Skip to Content (Press Enter)
      </button>

      <div className={`app__sidebar ${mobileSidebarIsActive ? 'app__sidebar--is-visible' : ''}`}>
        <div className="app__header">
          <div>
            {config.org || 'DSK'} /{' '}
            <BaseLink
              router={props.router}
              routeName="home"
              routeParams={{ v: props.route.params.v }}
              className="app__title"
            >
              {config.project}
            </BaseLink>
          </div>
        </div>
        <div className="app__nav">
          <TreeNavigation
            tree={tree}
            hideMobileSidebar={() => {
              setMobileSidebarIsActive(false);
            }}
          />
        </div>
        <div className="app__shoutout">
          {availableSources && (
            <div className="app__versions">
              <select
                value={source}
                onChange={ev => {
                  changeSource(ev.target.value);
                }}
              >
                {availableSources.map(s => {
                  return (
                    <option key={s.name} value={s.name} disabled={!s.is_ready}>
                      Version: {s.name} {s.is_ready ? '' : '(loading...)'}
                    </option>
                  );
                })}
              </select>
            </div>
          )}
          Powered by <a href="https://github.com/atelierdisko/dsk">DSK</a> ·{' '}
          <a href="mailto:thankyou@rundsk.com">Get in Touch</a>
        </div>
      </div>
      <main className="app__main" ref={refToMain} tabIndex="0">
        <div className="app__mobile-header">
          <div
            className="app__mobile-header-icon"
            onClick={() => {
              setMobileSidebarIsActive(!mobileSidebarIsActive);
            }}
          >
            {mobileSidebarIsActive ? (
              <img src={CloseIcon} alt="Toggle Menu" />
            ) : (
              <img src={HamburgerIcon} alt="Toggle Menu" />
            )}
          </div>
          <div>
            {config.org || 'DSK'} /{' '}
            <BaseLink
              router={props.router}
              routeName="home"
              routeParams={{ v: props.route.params.v }}
              className="app__title"
            >
              {config.project}
            </BaseLink>
          </div>
        </div>

        {content}
      </main>
      <div className="app__search">
        <Search title={config.project} />
      </div>
    </div>
  );
}

export default routeNode('')(App);
