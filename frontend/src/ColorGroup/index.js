import React, { useState, useEffect } from 'react';
import './ColorGroup.css';
import { Client } from '@atelierdisko/dsk';
import ColorCard from '../ColorCard';

function ColorGroup(props) {
  const [colors, setColors] = useState([]);

  useEffect(() => {
    getData();
  }, [props.src]);

  function getData() {
    if (props.src) {
      // FIXME: This should obviously be derived from the src attribute
      Client.get(`/Basics/Colors${props.src.slice(1)}`).then(data => {
        setColors(data.colors);
      });
    }
  }

  let classes = ['color-group'];

  let content = props.children;

  if (props.compact) {
    classes.push('color-group--is-compact');

    // We have to make sure the compact property is set on all
    // children as well
    content = React.Children.map(props.children, c => {
      return React.cloneElement(c, { compact: true });
    });
  }

  // If the src prop is set, the information about the colors should
  // be loaded via the API
  if (props.src) {
    content = colors.map(c => {
      return (
        <ColorCard key={c.value} color={c.value} comment={c.comment} id={c.id} compact={props.compact}>
          {c.name}
        </ColorCard>
      );
    });
  }

  return <div className={classes.join(' ')}>{content}</div>;
}

export default ColorGroup;
