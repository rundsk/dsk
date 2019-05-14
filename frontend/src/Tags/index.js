import React, { useState, useEffect } from 'react';
import { useGlobal } from 'reactn';
import contrast from 'get-contrast';

import './Tags.css';

function Tags(props) {
  const [style, setStyle] = useState("");
  const [filterTerm, setFilterTerm] = useGlobal("filterTerm");
  const [frontendConfig] = useGlobal("frontendConfig");

  useEffect(() => {
    if (frontendConfig.tags) {
      let styles = frontendConfig.tags.map((t) => {

        let textColor = "rgba(0,0,0,.7)";

        if (!t.color.includes("var") && contrast.ratio(t.color, "black") < contrast.ratio(t.color, "white")) {
          textColor = "white";
        }

        if (t.color.includes("--color-red") || t.color.includes("--color-blue")) {
          textColor ="white";
        }

        return `.tags li.tags__tag[tag*='${t.name}'] { background-color: ${t.color}; color: ${textColor}; }`;
      });
      setStyle(styles.join("\n"));
    }
  }, [frontendConfig]);

  let tags;

  if (props.tags) {
    tags = props.tags.map((t) => {
      return <li
        className="tags__tag"
        key={t}
        tag={t}
        onClick={(ev) => {
          if (props.isClickable === false) { return; }

          if (ev.metaKey) {
            setFilterTerm(`${filterTerm} ${t}`);
          } else {
            setFilterTerm(t);
          }
        }}
      >{t}</li>;
    })
  }

  return (
    <ul className="tags">
      <style>{style}</style>

      {tags}
    </ul>
  );
}

export default Tags;