/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useEffect, useRef, useMemo, useCallback, useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { Helmet } from 'react-helmet';

import { constructURL, slugify } from '../utils';

import Breadcrumbs from '../Breadcrumbs';
import Tags from '../Tags';
import Playground from '../DocumentationComponents/Playground';
import Doc from '../Doc';
import Meta from '../Meta';
import TabBar from '../TabBar';
import AssetList from '../AssetList';
import SourceView from '../SourceView';
import NodeList from '../NodeList';
import Link from '../Link';

import './Page.css';
import { GlobalContext } from '../App';

function scrollHeadingFromURLIntoView(activeTabWithHeading, docRef, smoothScroll = true) {
  const h = activeTabWithHeading?.split('§')[1] || '';

  if (h !== '' && docRef?.current) {
    let heading = docRef.current.querySelector(`[heading-id='${h}']`);

    if (heading) {
      heading.scrollIntoView({ behavior: smoothScroll ? 'smooth' : 'auto', block: 'start' });
    }
  }
}

function Page({
  url,
  id,
  title,
  description,
  version,
  modified,
  crumbs,
  prev,
  next,
  tags,
  related,
  authors,
  custom,
  docs,
  downloads,
  source,
  activeTab: activeTabWithHeading,
  children,
}) {
  const history = useHistory();
  const { config } = useContext(GlobalContext);
  const docRef = useRef(null);

  // Scroll to top on navigation
  useEffect(() => {
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto',
    });
  }, [url]);

  // When the Tab part of the URL changes we check if it contains
  // a heading id and if yes scroll there.
  useEffect(() => {
    scrollHeadingFromURLIntoView(activeTabWithHeading, docRef);
  }, [activeTabWithHeading]);

  // We also want to this once the Doc finished rendering for the first time.
  // We have to keep the function in a ref so a change of "activeTabWithHeading"
  // does not trigger a rerender of the Document
  let handleHeadingScrollRef = useRef();
  useEffect(() => {
    handleHeadingScrollRef.current = () => {
      scrollHeadingFromURLIntoView(activeTabWithHeading, docRef, false);
    };
  }, [activeTabWithHeading]);
  const handleDocOnRender = useCallback(() => {
    if (!handleHeadingScrollRef.current) {
      return;
    }
    handleHeadingScrollRef.current();
  }, []);

  const nodeInfo = useMemo(() => {
    return { id: id, url: url, title: title };
  }, [id, url, title]);

  function navigateToActiveTab(t) {
    let nextURL = constructURL({ activeTab: slugify(t) });
    history.replace(nextURL);
  }

  let playground;
  let tabBar;
  let doc;

  if (docs) {
    let rightSideTabs = [];

    docs = docs.filter((d) => {
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
      children &&
      children.length > 0 &&
      docs.filter((doc) => {
        return doc.title.toLowerCase() !== 'authors';
      }).length === 0;

    if (showOverview) {
      docs.push({
        title: 'Overview',
        content: <NodeList nodes={children} source={source} />,
      });
    }

    if (downloads && downloads.length > 0) {
      rightSideTabs.push({
        title: 'Assets',
        content: <AssetList assets={downloads} source={source} />,
      });
    }

    if (url === '') {
      // On the root node we also want to display the DSK version
      rightSideTabs.push({
        title: 'Source',
        content: (
          <>
            <SourceView url={'hello'} />
            <SourceView url={url} source={source} />
          </>
        ),
      });
    } else {
      rightSideTabs.push({
        title: 'Source',
        content: <SourceView url={url} source={source} />,
      });
    }

    // We find the active tab by removing the section part from the prop
    const activeTab = activeTabWithHeading?.split('§')[0];

    tabBar = (
      <TabBar
        onSetActiveTab={navigateToActiveTab}
        activeTab={activeTab}
        tabs={docs.map((d) => d.title)}
        rightSideTabs={rightSideTabs.map((d) => d.title)}
      />
    );

    let activeDoc = [...docs, ...rightSideTabs].find((d) => {
      return slugify(d.title) === activeTab;
    });

    if (!activeDoc && docs.length > 0) {
      activeDoc = docs[0];
    }

    // If this is not an overview/asset/source doc, its content comes
    // from the API in the form of HTML, rather than React elements
    if (activeDoc && activeDoc.html) {
      doc = (
        <Doc
          id={activeDoc.id}
          url={activeDoc.url}
          title={activeDoc.title}
          htmlContent={activeDoc.html}
          toc={activeDoc.toc}
          components={activeDoc.components}
          node={nodeInfo}
          onRender={handleDocOnRender}
        />
      );
    }

    if (activeDoc && activeDoc.content) {
      doc = (
        <Doc
          id={activeDoc.id}
          url={activeDoc.url}
          title={activeDoc.title}
          toc={activeDoc.toc}
          components={activeDoc.components}
          node={nodeInfo}
          onRender={handleDocOnRender}
        >
          {activeDoc.content}
        </Doc>
      );
    }
  }

  if (authors && authors.length > 0) {
    let title = authors.length > 1 ? 'Authors' : 'Author';
    let authorLinks = authors.map((a, i) => {
      return (
        <span className="page__author" key={a.email}>
          {i > 0 ? ', ' : ''}
          <a href={`mailto:${a.email}`}>{a.name !== '' ? a.name : a.email}</a>
        </span>
      );
    });

    authors = <Meta title={title}>{authorLinks}</Meta>;
  }

  if (related && related.length > 0) {
    let title = 'Related';
    let relatedLinks = related.map((a, i) => {
      return (
        <span className="page__related" key={a.url}>
          {i > 0 ? ', ' : ''}

          <Link to={a.url}>{a.title}</Link>
        </span>
      );
    });

    related = <Meta title={title}>{relatedLinks}</Meta>;
  }

  if (custom) {
    // Turn props.custom object into array to be able to iterate with map()
    custom = Object.entries(custom).map((data) => {
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

  const modifiedFormatted = useMemo(() => new Date(modified * 1000).toLocaleDateString(), [modified]);

  return (
    <div className="page">
      <Helmet
        titleTemplate={`%s – ${config?.org} / ${config?.project}`}
        defaultTitle={`${config?.org} / ${config?.project}`}
      >
        <title>{url !== '' && title}</title>
        <meta name="description" content={description} />
      </Helmet>
      <div className="page__header">
        <div className="page__header-content">
          <Breadcrumbs crumbs={crumbs} />

          <h1 className="page__title">{title}</h1>
          <p className="page__description">
            {description}
            <span className="page__children-count">
              {children && children.length > 0 && ` (${children.length} aspects)`}
            </span>
          </p>

          <Tags tags={tags} />

          <div className="page__meta">
            <div className="page__meta-items-container">
              <div className="page__meta-item">
                <Meta title="Last Changed">{modifiedFormatted}</Meta>
              </div>
              {version && (
                <div className="page__meta-item">
                  <Meta title="Version">{version}</Meta>
                </div>
              )}
              {authors && <div className="page__meta-item">{authors}</div>}
              {related && <div className="page__meta-item">{related}</div>}
              {custom}
            </div>
          </div>
        </div>
      </div>

      {playground && (
        <div className="page__component-demo">
          <Playground isPageComponentDemo contentFullWidth>
            <Doc htmlContent={playground.html} id="playground" node={nodeInfo} />
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
          {prev && (
            <Link to={`/${prev.url}`} className="page__node-nav page__node-nav--prev">
              <Meta title="Previous">{prev.title}</Meta>
            </Link>
          )}

          {next && (
            <Link to={`/${next.url}`} className="page__node-nav page__node-nav--next">
              <Meta title="Next">{next.title}</Meta>
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}

export default React.memo(Page);
