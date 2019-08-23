/**
 * Copyright 2019 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

import React from 'react';
import ReactDOM from 'react-dom';
import CodeBlock from '.';
import { shallow } from 'enzyme';
import 'jest-enzyme';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<CodeBlock />, div);
});

it('wraps content in code block', () => {
  const component = shallow(<CodeBlock>Hello World!</CodeBlock>);
  const code = <code className="code-block__code-content">Hello World!</code>;
  expect(component).toContainReact(code);
});

it('escapes HTML content', () => {
  const HTML = '<button>Fancy</button>';
  const component = shallow(<CodeBlock>{HTML}</CodeBlock>);

  expect(component.find('code').html()).toEqual(
    '<code class="code-block__code-content">&lt;button&gt;Fancy&lt;/button&gt;</code>'
  );
});

it('does not escape pre-escaped HTML content', () => {
  const HTML = '&lt;button&gt;Fancy&lt;/button&gt;';
  const component = shallow(<CodeBlock escaped>{HTML}</CodeBlock>);

  expect(component.find('code').html()).toEqual(
    '<code class="code-block__code-content">&lt;button&gt;Fancy&lt;/button&gt;</code>'
  );
});

it('renders pre-escaped content with initial blank line', () => {
  const HTML = `
authors:
  - christoph@atelierdisko.de
  - marius@atelierdisko.de

description: &gt;
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.
`;
  const component = shallow(<CodeBlock escaped>{HTML}</CodeBlock>);

  const expected = `authors:
  - christoph@atelierdisko.de
  - marius@atelierdisko.de

description: &gt;
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.
`;
  expect(component.find('code').html()).toEqual(
    `<code class="code-block__code-content">${expected}</code>`
  );
});

it('renders component build up content', () => {
  const component = shallow(
    <CodeBlock title="Example">
      <div>test</div>
    </CodeBlock>
  );

  const expected = `&lt;div&gt;test&lt;/div&gt;`;

  expect(component.find('code').html()).toEqual(
    `<code class="code-block__code-content">${expected}</code>`
  );
});
