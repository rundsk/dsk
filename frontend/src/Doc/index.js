/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useEffect } from 'react';
import { transform } from '@rundsk/js-sdk';
import { withRoute } from 'react-router5';

import './Doc.css';

import Heading from '../Heading';
import Image from '../Image';
import Link from '../Link';

import Banner from '../Banner';
import CodeBlock from '../CodeBlock';
import Color from '../Color';
import ColorCard from '../ColorCard';
import ColorGroup from '../ColorGroup';
import Playground from '../Playground';
import DoDont, { Do, Dont } from '../DoDont';
import FigmaEmbed from '../FigmaEmbed';
import TypographySpecimen from '../TypographySpecimen';
import Glitch from '../Glitch';
import CodeSandbox from '../CodeSandbox';
import Asciinema from '../Asciinema';
import ImageGrid from '../ImageGrid';
import TableOfContents from '../TableOfContents';

function Doc(props) {
  useEffect(() => {
    // window.requestAnimationFrame should ensure that the rendering
    // has finished.
    window.requestAnimationFrame(() => {
      if (props.onRender) {
        props.onRender();
      }
    });
  });

  if (!props.htmlContent) {
    return <div className="doc">{props.children}</div>;
  }

  // Transform context: Allow to use this inside the `transforms` constant. We
  // cannot use `props.x` there as that refers to the component/element that is
  // being transformed.
  let context = {
    node: props.node,
    doc: { id: props.id, url: props.url, title: props.title },
  };

  const transforms = {
    Banner: (props) => <Banner {...props} />,
    CodeBlock: (props) => {
      // When using <CodeBlock> directly within documents, its contents aren't
      // automatically protected from interpration as HTML, when they processed
      // by the DocTransformer. Thus we expect users to wrap their literal code
      // in <script> tags, which we again remove here.
      let children = props.children.replace(/^\s*<script>/, '').replace(/<\/script>\s*$/, '');

      return <CodeBlock {...props} {...context} children={children} />;
    },
    ColorCard: (props) => <ColorCard {...props} {...context} />,
    ColorGroup: (props) => <ColorGroup {...props} {...context} />,
    Color: (props) => <Color {...props} {...context} />,
    Playground: (props) => <Playground {...props} {...context} />,
    Do: (props) => <Do {...props} {...context} />,
    DoDontGroup: (props) => <DoDont {...props} {...context} />,
    Dont: (props) => <Dont {...props} {...context} />,
    FigmaEmbed: (props) => <FigmaEmbed {...props} {...context} />,
    Image: (props) => <Image {...props} {...context} />,
    TypographySpecimen: (props) => <TypographySpecimen {...props} {...context} />,
    Warning: (props) => <Banner type="warning" {...props} {...context} />,
    Glitch: (props) => <Glitch {...props} {...context} />,
    CodeSandbox: (props) => <CodeSandbox {...props} {...context} />,
    Asciinema: (props) => <Asciinema {...props} {...context} />,
    ImageGrid: (props) => <ImageGrid {...props} {...context} />,
    TableOfContents: (props) => <TableOfContents {...props} {...context} />,

    a: (props) => <Link {...props} {...context} />,

    h1: (props) => <Heading {...props} {...context} level="alpha" isJumptarget={true} />,
    h2: (props) => <Heading {...props} {...context} level="beta" isJumptarget={true} />,
    h3: (props) => <Heading {...props} {...context} level="gamma" isJumptarget={true} />,
    h4: (props) => <Heading {...props} {...context} level="delta" isJumptarget={true} />,

    img: (props) => <Image {...props} {...context} />,
    pre: (props) => {
      // When a language is added to a Markdown fenced code block, it is
      // stored as a class with a "language-" prefix in the inner <code>.
      // Here we extract it and turn it into a prop.
      let language = props.children.match(/^<code class="language-(.*?)">/);

      if (language && language.length === 2) {
        props['language'] = language[1];
      }

      // When Markdown fenced code blocks get converted to <pre> they
      // additionally include inner <code>. We cannot use orphans as this create
      // empty "ghost" elements.
      let children = props.children.replace(/^<code>/, '').replace(/<\/code>$/, '');

      return <CodeBlock escaped {...props} {...context} children={children} />;
    },
  };

  const orphans = Object.keys(transforms)
    .filter((k) => k !== 'a')
    .filter((k) => k !== 'Color')
    .map((k) => `p > ${k}`)
    .concat(['p > video']);

  let transformedContent = transform(props.htmlContent, transforms, orphans, {
    isPreformatted: (type) => ['pre', 'CodeBlock', 'Playground'].map((v) => v.toLowerCase()).includes(type),
    noTransform: (type, props) => {
      // This gets called on HTML elements that do not need
      // to be transformed to special React components.
      // There are differences between the attributes of
      // HTML elements and React that we have to take care
      // of: https://reactjs.org/docs/dom-elements.html#differences-in-attributes
      props.className = props.class;
      delete props.class;

      return React.createElement(type, props, props.children);
    },
  });
  return <div className="doc">{transformedContent}</div>;
}

export default withRoute(Doc);
