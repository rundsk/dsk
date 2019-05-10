import React from 'react';
import './Meta.css';

function Meta(props) {
  return (
    <div className="meta">
      <div className="meta__key">{props.title}</div>
      <div className="meta__value">{props.children}</div>
    </div>
  );
}

export default Meta;
