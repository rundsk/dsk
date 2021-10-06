/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2018 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React from 'react';
import './AssetList.css';
import Meta from '../Meta';
import Heading from '../DocumentationComponents/Heading';
import filesize from 'filesize';

function AssetList(props) {
  const imageFileTypes = ['png', 'jpg', 'jpeg'];

  return (
    <>
      {props.assets &&
        props.assets.map((a) => {
          return (
            <div className="asset-list__asset" key={a.name}>
              <Heading level="beta" isJumptarget={true} docTitle="Assets">
                {a.name}
              </Heading>

              {imageFileTypes.some((v) => {
                return a.url.indexOf(v) >= 0;
              }) && (
                <img className="asset-list__asset-image" src={`/api/v1/tree/${a.url}?v=${props.source}`} alt={a.name} />
              )}

              <div className="asset-list__asset-meta">
                <Meta title="Download">
                  <a className="asset-list__asset-download" href={`/api/v1/tree/${a.url}?v=${props.source}`} download>
                    {a.name}
                  </a>
                </Meta>
                {a.url.indexOf('json') >= 0 && (
                  <Meta title="Download (Converted)">
                    <a
                      className="asset-list__asset-download"
                      href={`/api/v1/tree/${a.url.replace('json', 'yaml')}?v=${props.source}`}
                      download
                    >
                      {a.name.replace('json', 'yaml')}
                    </a>
                  </Meta>
                )}
                {a.url.indexOf('yaml') >= 0 && (
                  <Meta title="Download (Converted)">
                    <a
                      className="asset-list__asset-download"
                      href={`/api/v1/tree/${a.url.replace('yaml', 'json')}?v=${props.source}`}
                      download
                    >
                      {a.name.replace('yaml', 'json')}
                    </a>
                  </Meta>
                )}
                <Meta title="Size">{filesize(a.size)}</Meta>
                {a.width && a.height && (
                  <Meta title="Dimensions">
                    {a.width}px × {a.height}px
                  </Meta>
                )}
                <Meta title="Last Modified">{new Date(a.modified * 1000).toLocaleString()}</Meta>
              </div>
            </div>
          );
        })}
    </>
  );
}

export default AssetList;
