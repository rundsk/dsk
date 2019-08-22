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

it('does escape pre-escaped HTML content', () => {
  const HTML = '&lt;button&gt;Fancy&lt;/button&gt;';
  const component = shallow(<CodeBlock escaped>{HTML}</CodeBlock>);

  expect(component.find('code').html()).toEqual(
    '<code class="code-block__code-content">&lt;button&gt;Fancy&lt;/button&gt;</code>'
  );
});
