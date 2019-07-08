/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from "react";
import { BaseLink, withRoute } from "react-router5";
import "./Breadcrumbs.css";

function Breadcrumbs(props) {
  let crumbs;

  if (props.crumbs) {
    crumbs = props.crumbs.map(c => {
      return (
        <li className="breadcrumbs__crumb" key={c.title}>
          <BaseLink router={props.router} routeName="node" routeParams={{ node: `${c.url}` }} key={"link"}>
            {c.title}
          </BaseLink>
          {/* <a href={`/tree/${c.url}`}>{c.title}</a> */}
        </li>
      );
    });

    crumbs.pop();
  }
  return <ul className="breadcrumbs">{crumbs}</ul>;
}

export default withRoute(Breadcrumbs);
