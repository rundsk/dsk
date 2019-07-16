/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import { transform } from '@atelierdisko/dsk';
import { withRoute } from 'react-router5';

import './Doc.css';

import Heading from '../Heading';
import Image from '../Image';
import Link from '../Link';

import Banner from '../Banner';
import CodeBlock from '../CodeBlock';
import ColorCard from '../ColorCard';
import ColorGroup from '../ColorGroup';
import ComponentDemo from '../ComponentDemo';
import DoDont, { Do, Dont } from '../DoDont';
import FigmaEmbed from '../FigmaEmbed';
import TypographySpecimen from '../TypographySpecimen';

function Doc(props) {
  if (!props.htmlContent) {
    return <div className="doc">{props.children}</div>;
  }

  // Allow to use this inside the `transforms` constant. We cannot use
  // `props.title` there as that refers to the component/element that is being
  // transformed. The title is needed as to calculate the `Heading`'s jump
  // anchor.
  let docTitle = props.title;

  const transforms = {
    Banner: props => <Banner {...props} />,
    CodeBlock: props => <CodeBlock {...props} />,
    Color: props => <ColorCard {...props} />,
    ColorGroup: props => <ColorGroup {...props} />,
    Playground: props => <ComponentDemo {...props} />,
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
    pre: props => <CodeBlock {...props} />,
  };

  const orphans = ['p > img', 'p > video'];

  let transformedContent = transform(props.htmlContent, transforms, orphans, {
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
