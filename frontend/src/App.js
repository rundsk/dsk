import React, { useState, useEffect } from 'react';
import { routeNode, BaseLink } from 'react-router5'
import { useGlobal } from 'reactn';

import { Client } from '@atelierdisko/dsk';
import TreeNavigation from './TreeNavigation';

import './Variables.css';
import './App.css';
import Page from './Page';
import ErrorPage from './ErrorPage';
import Search from './Search';

function App(props) {
  const [tree, setTree] = useState(null);
  const [title, setTitle] = useState("Design System");
  const [node, setNode] = useState(null);
  const [currentNode, setCurrentNode] = useState(null);
  const [error, setError] = useState(null);
  const [frontendConfig, setFrontendConfig] = useGlobal("frontendConfig");

  function getNode() {
    let nodeToGet = currentNode;
    if (currentNode === null) { return; }
    if (currentNode === "root") { nodeToGet = "" };

    Client.get(nodeToGet).then((data) => {
      setNode(data);
      setError(null)
    }).catch((err)  =>{
      console.log(err);
      setError("Node not found.");
    });
  }

  useEffect(() => {
    Client.tree().then((data) => {
      setTree(data.root);
      setTitle(data.root.title)
    }).catch((err) => {
      console.log(err);
    });

    Client.get("/frontendConfig.some").then((data) => {
      setFrontendConfig(data);
    }).catch((err) => {
      console.log("No config found" + err);
    });
  }, []);

  useEffect(() => {
    switch (props.route.name) {
      case "home":
        setCurrentNode("root");
        break;
      case "node":
        setCurrentNode(props.route.params.node);
        break;
      default:
        break;
    }
  }, [props.route]);

  useEffect(() => {
    getNode();
  }, [currentNode]);

  let content;
  if (node) {
    let activeTab = props.route.params.t || undefined;
    content = <Page {...node} activeTab={activeTab} designSystemTitle={title} />;
  }

  if (error) {
    content = <ErrorPage>{error}</ErrorPage>;
  }

  let refToMain = React.createRef();

  return (
    <div className="app">
      <button className="app_skip-to-content" onClick={() => { if (refToMain.current) { console.log(refToMain); refToMain.current.focus() } }}>Skip to Content (Press Enter)</button>

      <div className="app__sidebar">
        <div className="app__header">
          <div>{frontendConfig.organisation || "DSK"} / <BaseLink router={props.router} routeName="home" className="app__title">{title}</BaseLink></div>
        </div>
        <div className="app__nav">
          <TreeNavigation tree={tree} />
        </div>
        <div className="app__shoutout">
          Powered by <a href="https://github.com/atelierdisko/dsk">DSK</a> · <a href="mailto:designsystems@atelierdisko.de">Get in Touch</a>
        </div>
      </div>
      <main className="app__main" ref={refToMain} tabIndex="0">
        {content}
      </main>
      <div className="app__search"><Search title={title} /></div>
    </div>
  );
}

export default routeNode('')(App)
