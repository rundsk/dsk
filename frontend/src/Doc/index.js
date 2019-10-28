/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React, { useEffect } from 'react';
import useScript, { ScriptStatus } from '@charlietango/use-script'
import { transform } from '@atelierdisko/dsk';
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

function Doc(props) {
  const [ready, status] = useScript('/api/v1/tree/bundles/example.js');

  console.log('ready?', ready);
  if (ready) {
    var myButton = window.System.load('Button');
    console.log(myButton);
  }

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

  // Allow to use this inside the `transforms` constant. We cannot use
  // `props.title` there as that refers to the component/element that is being
  // transformed. The title is needed as to calculate the `Heading`'s jump
  // anchor.
  let docTitle = props.title;

  const transforms = {
    MyButton: (props) => {
      if (myButton) {
        return myButton(props);
      }
    },
    Banner: props => <Banner {...props} />,
    CodeBlock: props => {
      // When using <CodeBlock> directly within documents, its contents aren't
      // automatically protected from interpration as HTML, when they processed
      // by the DocTransformer. Thus we expect users to wrap their literal code
      // in <script> tags, which we again remove here.
      let children = props.children.replace(/^\s*<script>/, '').replace(/<\/script>\s*$/, '');

      return <CodeBlock {...props} children={children} />;
    },
    ColorCard: props => <ColorCard {...props} />,
    ColorGroup: props => <ColorGroup {...props} />,
    Color: props => <Color {...props} />,
    Playground: props => <Playground {...props} />,
    Do: props => <Do {...props} />,
    DoDontGroup: props => <DoDont {...props} />,
    Dont: props => <Dont {...props} />,
    FigmaEmbed: props => <FigmaEmbed {...props} />,
    TypographySpecimen: props => <TypographySpecimen {...props} />,
    Warning: props => <Banner type="warning" {...props} />,

    a: props => <Link {...props} />,
    h1: props => <Heading {...props} level="alpha" docTitle={docTitle} isJumptarget={true} />,
    h2: props => <Heading {...props} level="beta" docTitle={docTitle} isJumptarget={true} />,
    h3: props => <Heading {...props} level="gamma" docTitle={docTitle} isJumptarget={true} />,
    h4: props => <Heading {...props} level="delta" docTitle={docTitle} isJumptarget={true} />,
    img: props => <Image {...props} />,
    pre: props => {
      // When Markdown fenced code blocks get converted to <pre> they
      // additionally include inner <code>. We cannot use orphans as this create
      // empty "ghost" elements.
      let children = props.children.replace(/^<code>/, '').replace(/<\/code>$/, '');

      return <CodeBlock escaped {...props} children={children} />;
    },
  };

  const orphans = Object.keys(transforms)
    .filter(k => k !== 'a')
    .filter(k => k !== 'Color')
    .map(k => `p > ${k}`)
    .concat(['p > video']);

  let transformedContent = transform(props.htmlContent, transforms, orphans, {
    isPreformatted: type => type === 'pre' || type === 'CodeBlock'.toLowerCase(),
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
