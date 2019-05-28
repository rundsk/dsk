import React, { useEffect, useRef } from 'react';
import { BaseLink, withRoute } from 'react-router5';
import { slugify } from '../utils';

import './Page.css';

import Breadcrumbs from '../Breadcrumbs';
import Tags from '../Tags';
import ComponentDemo from '../ComponentDemo';
import Doc from '../Doc';
import Meta from '../Meta';
import TabBar from '../TabBar';
import AssetList from '../AssetList';
import SourceView from '../SourceView';
import NodeList from '../NodeList';

function Page(props) {
  const docRef = useRef(null);

  useEffect(() => {
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto'
    });
  }, [props.url]);

  useEffect(() => {
    let title = `${props.designSystemTitle}: ${props.title}`;

    if (props.url === "") {
      title = `${props.designSystemTitle}`;
    }

    document.title = title;
  }, [props.title, props.designSystemTitle]);

  function navigateToActiveTab(t) {
    // We handle tab selection completly via the URL
    let currentRouterState = props.router.getState();

    // On root there is no node parameter
    let currentNode = currentRouterState.params.node || "";
    t = slugify(t);
    props.router.navigate("node", { ...currentRouterState.params, node: currentNode, t: t }, { replace: true });
  }

  useEffect(() => {
    // FIXME: We delay this, because we hope by the time
    // we call the function all the children have loaded
    // their async content and have reached their final 
    // size. There is probably a better way to do this.
    setTimeout(() => {
      let currentRouterState = props.router.getState();
      let h = currentRouterState.params.t || "";
      h = h.split("§")[1] || "";
      if (h !== "" && docRef.current) {
        docRef.current.querySelector(`[heading-id='${h}']`).scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }, 300);
  });

  let componentDemo;
  let tabBar;
  let doc;

  if (props.docs) {
    let docs = [];
    let rightSideTabs = [];

    docs = props.docs.filter((d) => {
      if (d.title.toLowerCase() === "componentdemo") {
        componentDemo = d;
        return false;
      }

      return true;
    });

    // Pages without docs and root show an overview
    // of their children
    if ((docs.length === 0 || props.url === "") && props.children.length > 0) {
      docs.unshift({
        title: "Overview",
        content: <Doc title="Overview"><NodeList nodes={props.children} /></Doc>
      });
    }

    if (props.downloads && props.downloads.length > 0) {
      rightSideTabs.push({
        title: "Assets",
        content: <Doc title="Assets"><AssetList assets={props.downloads} /></Doc>
      });
    }

    rightSideTabs.push({
      title: "Source",
      content: <Doc title="Source"><SourceView url={props.url} /></Doc>
    });

    // We find the active tab by removing the section part from the prop
    let activeTab = props.activeTab;
    if (activeTab) { activeTab = activeTab.split("§")[0] }
    if (activeTab === "") { activeTab = undefined; }

    tabBar = <TabBar onSetActiveTab={navigateToActiveTab} activeTab={activeTab} tabs={docs.map(d => d.title)} rightSideTabs={rightSideTabs.map(d => d.title)} />

    let activeDoc = [...docs, ...rightSideTabs].find((d) => {
      return slugify(d.title) === activeTab;
    });

    if (!activeDoc && docs.length > 0) {
      activeDoc = docs[0];
    }

    if (activeDoc && activeDoc.html) {
      doc = <Doc title={activeDoc.title} content={activeDoc.html} />
    }

    if (activeDoc && activeDoc.content) {
      doc = <Doc title={activeDoc.title}>{activeDoc.content}</Doc>;
    }
  }

  let authors;

  if (props.authors && props.authors.length > 0) {
    let title = props.authors.length > 1 ? "Authors" : "Author";
    let authorLinks = props.authors.map((a, i) => {
      return <span className="page__author" key={a.email}>{i > 0 ? ", " : ""}<a href={`mailto:${a.email}`}>{a.name !== "" ? a.name : a.email}</a></span>;
    })

    authors = <Meta title={title}>{authorLinks}</Meta>;
  }

  // Custom meta data
  let custom;

  if (props.custom) {
    // Turn props.custom object into array to be able to iterate with map()
    custom = Object.entries(props.custom).map(data => {
      let title = data[0];

      // Display value list or single value
      let value;
      if (Array.isArray(data[1])) {
        value = data[1].join(", ");
      } else {
        value = data[1];
      }

      return <Meta key={title} title={title}>{value}</Meta>
    });
  }

  return (
    <div className="page">
      <div className="page__header">
        <div className="page__header-content">
          <Breadcrumbs crumbs={props.crumbs} />

          <h1 className="page__title">{props.title}</h1>
          <p className="page__description">
            {props.description}
            <span className="page__children-count">{props.children.length > 0 && ` (${props.children.length} aspects)`}</span>
          </p>

          <Tags tags={props.tags} />

          <div className="page__meta">
            <Meta title="Last Changed">{new Date(props.modified * 1000).toLocaleDateString()}</Meta>
            {authors}
            {custom}
          </div>
        </div>
      </div>

      {componentDemo &&
        <div className="page__component-demo">
          <ComponentDemo>
            <Doc content={componentDemo.html} />
          </ComponentDemo>
        </div>
      }

      {tabBar &&
        <div className="page__tabs">
          <div className="page__tabs-content">
            {tabBar}
          </div>
        </div>
      }

      <div className="page__docs" ref={docRef}>
        {doc}
      </div>

      <div className="page__footer">
        <div className="page__footer-content">
          {props.prev &&
            <BaseLink router={props.router} routeName='node' routeParams={{ node: `${props.prev.url}` }} className="page__node-nav page__node-nav--prev">
              <Meta title="Previous">{props.prev.title}</Meta>
            </BaseLink>
          }

          {props.next &&
            <BaseLink router={props.router} routeName='node' routeParams={{ node: `${props.next.url}` }} className="page__node-nav page__node-nav--next">
              <Meta title="Next">{props.next.title}</Meta>
            </BaseLink>
          }
        </div>
      </div>
    </div>
  );
}

export default withRoute(Page);
