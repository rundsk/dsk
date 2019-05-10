import React from 'react';
import './Breadcrumbs.css';

function Breadcrumbs(props) {
  let crumbs;

  if (props.crumbs) {
    crumbs = props.crumbs.map((c) => {
      return <li className="breadcrumbs__crumb" key={c.title}><a href={`/tree/${c.url}`}>{c.title}</a></li>;
    })

    crumbs.pop();
  }
  return (
    <ul className="breadcrumbs">
      {crumbs}
    </ul>
  );
}

export default Breadcrumbs;
