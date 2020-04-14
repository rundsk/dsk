/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useEffect, useRef } from 'react';
import { BaseLink, withRoute } from 'react-router5';
import { slugify } from '../utils';
import { Helmet } from 'react-helmet';

import './Page.css';

import Breadcrumbs from '../Breadcrumbs';
import Tags from '../Tags';
import Playground from '../Playground';
import Doc from '../Doc';
import Meta from '../Meta';
import TabBar from '../TabBar';
import AssetList from '../AssetList';
import SourceView from '../SourceView';
import NodeList from '../NodeList';

function Page(props) {
  const docRef = useRef(null);

  // Scroll to top on navigation
  useEffect(() => {
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto',
    });
  }, [props.url]);

  function docDidRender() {
    // Check if there is a section marker in the URL, and if their is
    // scroll there
    let currentRouterState = props.router.getState();
    let h = currentRouterState.params.t || '';
    h = h.split('§')[1] || '';
    if (h !== '' && docRef.current) {
      docRef.current.querySelector(`[heading-id='${h}']`).scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }

  function navigateToActiveTab(t) {
    // We handle tab selection completly via the URL
    let currentRouterState = props.router.getState();

    // On root there is no node parameter
    let currentNode = currentRouterState.params.node || '';
    t = slugify(t);
    props.router.navigate('node', { ...currentRouterState.params, node: currentNode, t: t }, { replace: true });
  }

  let playground;
  let tabBar;
  let doc;

  if (props.docs) {
    let docs = [];
    let rightSideTabs = [];

    docs = props.docs.filter(d => {
      if (d.title.toLowerCase() === 'playground') {
        playground = d;
        return false;
      }

      return true;
    });

    // Pages without `docs` show an overview of their children. Most DDTs will
    // have an `AUTHORS.txt` (which is included in `docs`), but some may lack
    // any "real" document that can be presented to the user. In this case we
    // also want to show an overview.
    const showOverview =
      props.children &&
      props.children.length > 0 &&
      docs.filter(doc => {
        return doc.title.toLowerCase() !== 'authors';
      }).length === 0;

    if (showOverview) {
      docs.push({
        title: 'Overview',
        content: <NodeList nodes={props.children} source={props.source} />,
      });
    }

    if (props.downloads && props.downloads.length > 0) {
      rightSideTabs.push({
        title: 'Assets',
        content: <AssetList assets={props.downloads} source={props.source} />,
      });
    }

    if (props.url === '') {
      // On the root node we also want to display the DSK version
      rightSideTabs.push({
        title: 'Source',
        content: (
          <>
            <SourceView url={'hello'} />
            <SourceView url={props.url} source={props.source} />
          </>
        ),
      });
    } else {
      rightSideTabs.push({
        title: 'Source',
        content: <SourceView url={props.url} source={props.source} />,
      });
    }

    // We find the active tab by removing the section part from the prop
    let activeTab = props.activeTab;
    if (activeTab) {
      activeTab = activeTab.split('§')[0];
    }
    if (activeTab === '') {
      activeTab = undefined;
    }

    tabBar = (
      <TabBar
        onSetActiveTab={navigateToActiveTab}
        activeTab={activeTab}
        tabs={docs.map(d => d.title)}
        rightSideTabs={rightSideTabs.map(d => d.title)}
      />
    );

    let activeDoc = [...docs, ...rightSideTabs].find(d => {
      return slugify(d.title) === activeTab;
    });

    if (!activeDoc && docs.length > 0) {
      activeDoc = docs[0];
    }

    // If this is not an overview/asset/source doc, its content comes
    // from the API in the form of HTML, rather than React elements
    if (activeDoc && activeDoc.html) {
      doc = <Doc title={activeDoc.title} htmlContent={activeDoc.html} onRender={docDidRender} />;
    }

    if (activeDoc && activeDoc.content) {
      doc = (
        <Doc title={activeDoc.title} onRender={docDidRender}>
          {activeDoc.content}
        </Doc>
      );
    }
  }

  let authors;

  if (props.authors && props.authors.length > 0) {
    let title = props.authors.length > 1 ? 'Authors' : 'Author';
    let authorLinks = props.authors.map((a, i) => {
      return (
        <span className="page__author" key={a.email}>
          {i > 0 ? ', ' : ''}
          <a href={`mailto:${a.email}`}>{a.name !== '' ? a.name : a.email}</a>
        </span>
      );
    });

    authors = <Meta title={title}>{authorLinks}</Meta>;
  }

  let related;

  if (props.related && props.related.length > 0) {
    let title = 'Related';
    let relatedLinks = props.related.map((a, i) => {
      return (
        <span className="page__related" key={a.url}>
          {i > 0 ? ', ' : ''}

          <BaseLink router={props.router} routeName="node" routeParams={{ node: `${a.url}`, v: props.route.params.v }}>
            {a.title}
          </BaseLink>
        </span>
      );
    });

    related = <Meta title={title}>{relatedLinks}</Meta>;
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
        value = data[1].join(', ');
      } else {
        value = data[1];
      }

      return (
        <div className="page__meta-item">
          <Meta key={title} title={title}>
            {value}
          </Meta>
        </div>
      );
    });
  }

  return (
    <div className="page">
      <Helmet titleTemplate={`%s – ${props.baseTitle}`} defaultTitle={props.baseTitle}>
        <title>{props.url !== '' && props.title}</title>
        <meta name="description" content={props.description} />
      </Helmet>
      <div className="page__header">
        <div className="page__header-content">
          <Breadcrumbs crumbs={props.crumbs} />

          <h1 className="page__title">{props.title}</h1>
          <p className="page__description">
            {props.description}
            <span className="page__children-count">
              {props.children && props.children.length > 0 && ` (${props.children.length} aspects)`}
            </span>
          </p>

          <Tags tags={props.tags} />

          <div className="page__meta">
            <div className="page__meta-items-container">
              <div className="page__meta-item">
                <Meta title="Last Changed">{new Date(props.modified * 1000).toLocaleDateString()}</Meta>
              </div>
              {props.version && <div className="page__meta-item">
                <Meta title="Version">{props.version}</Meta>
              </div>}
              {authors && <div className="page__meta-item">{authors}</div>}
              {related && <div className="page__meta-item">{related}</div>}
              {custom}
            </div>
          </div>
        </div>
      </div>

      {playground && (
        <div className="page__component-demo">
          <Playground isPageComponentDemo>
            <Doc htmlContent={playground.html} />
          </Playground>
        </div>
      )}

      {tabBar && (
        <div className="page__tabs">
          <div className="page__tabs-content">{tabBar}</div>
        </div>
      )}

      <div className="page__docs" ref={docRef}>
        {doc}
      </div>

      <div className="page__footer">
        <div className="page__footer-content">
          {props.prev && (
            <BaseLink
              router={props.router}
              routeName="node"
              routeParams={{ node: `${props.prev.url}`, v: props.route.params.v }}
              className="page__node-nav page__node-nav--prev"
            >
              <Meta title="Previous">{props.prev.title}</Meta>
            </BaseLink>
          )}

          {props.next && (
            <BaseLink
              router={props.router}
              routeName="node"
              routeParams={{ node: `${props.next.url}`, v: props.route.params.v }}
              className="page__node-nav page__node-nav--next"
            >
              <Meta title="Next">{props.next.title}</Meta>
            </BaseLink>
          )}
        </div>
      </div>
    </div>
  );
}

export default withRoute(Page);
