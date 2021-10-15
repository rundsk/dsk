/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect, useRef, useContext, useCallback } from 'react';
import { BrowserRouter as Router, Switch, Redirect, Route } from 'react-router-dom';
import Link from './Link';
import { Helmet } from 'react-helmet';

import { Client } from '@rundsk/js-sdk';

import Search from './Search';
import TreeNavigation from './TreeNavigation';
import SourcePicker from './SourcePicker';
import Node from './Node';

import HamburgerIcon from './HamburgerIcon.svg';
import CloseIcon from './CloseIcon.svg';

import './Variables.css';
import './App.css';

function Main() {
  const socket = useRef();

  const [tree, setTree] = useState(null);
  const [mobileSidebarIsActive, setMobileSidebarIsActive] = useState(false);

  const { config, setConfig, source } = useContext(GlobalContext);

  function loadTree(source) {
    if (!source) {
      return;
    }
    Client.tree(source)
      .then((data) => {
        setTree(data.root);
      })
      .catch((err) => {
        console.log(`Failed to load tree: ${err}`);
      });
  }

  const handleHideMobileSidebar = useCallback(() => {
    setMobileSidebarIsActive(false);
  }, [setMobileSidebarIsActive]);

  // Establish a WebSocket connection and register a handler. We're
  // intentionally not displaying notifications, as we consider them to be too
  // intrusive.
  useEffect(() => {
    if (socket.current) {
      // Ensure we don't open multiple sockets.
      return;
    }
    console.log('Establishing WebSocket connection...');
    socket.current = Client.messages();
    socket.current.addEventListener('message', (ev) => {
      // We relay the event via the window so we can listen to
      // it in multiple components that handle different concerns
      window.dispatchEvent(
        new CustomEvent('socketMessage', {
          detail: ev.data,
        })
      );
    });
  }, [socket]);

  // Handle Socket Event
  useEffect(() => {
    const handleSocketEvent = (ev) => {
      let m = JSON.parse(ev.detail);

      if (m.topic === `${source}.tree.synced`) {
        loadTree(source);
      }
    };
    window.addEventListener('socketMessage', handleSocketEvent);

    return () => {
      window.removeEventListener('socketMessage', handleSocketEvent);
    };
  }, [source]);

  // Load the config
  useEffect(() => {
    console.log('Loading Config');
    Client.config().then((data) => {
      setConfig({
        ...data,
      });
    });
  }, [setConfig]);

  // Initialize tree navigation.
  useEffect(() => {
    console.log('Loading the tree');
    loadTree(source);
  }, [source]);

  let refToMain = React.createRef();

  return (
    <div className="app">
      <Helmet htmlAttributes={{ lang: config?.lang }} />

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
            {config?.org || 'DSK'} /{' '}
            <Link to="/" className="app__title">
              {config?.project}
            </Link>
          </div>
        </div>
        <div className="app__nav">
          <TreeNavigation tree={tree} onHideMobileSidebar={handleHideMobileSidebar} />
        </div>
        <div className="app__shoutout">
          <SourcePicker />
          Powered by <a href="https://github.com/rundsk/dsk">DSK</a> ·{' '}
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
            {config?.org || 'DSK'} /{' '}
            <Link to="/" className="app__title">
              {config?.project}
            </Link>
          </div>
        </div>

        <Switch>
          <Route
            path="/tree/:node+"
            component={({ location }) => {
              // We redirect old URLs that still use the /tree prefix
              return <Redirect to={`${location.pathname.replace('/tree', '')}${location.search}`} />;
            }}
          ></Route>
          <Route path="/:node+">
            <Node />
          </Route>
          <Route path="/">
            <Node />
          </Route>
        </Switch>
      </main>
      <div className="app__search">
        <Search title={config?.project} />
      </div>
    </div>
  );
}

export const GlobalContext = React.createContext();

function App() {
  const [config, setConfig] = useState();
  const [source, setSource] = useState();
  const [filterTerm, setFilterTerm] = useState();

  return (
    <GlobalContext.Provider
      value={{
        config,
        setConfig,
        source,
        setSource,
        filterTerm,
        setFilterTerm,
      }}
    >
      <Router>
        <Main />
      </Router>
    </GlobalContext.Provider>
  );
}

export default App;
