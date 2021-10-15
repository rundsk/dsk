/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 *
 * @jest-environment jsdom
 */

import ReactDOM from 'react-dom';
import CodeBlock from '.';
import { mount } from 'enzyme';
import 'jest-enzyme';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<CodeBlock />, div);
});

it('wraps content in code block', () => {
  const component = mount(<CodeBlock>Hello World!</CodeBlock>);

  expect(component.find('code').html()).toEqual('<code class="code-block__code-content">Hello World!</code>');
});

it('does not escape pre-escaped HTML content', () => {
  const HTML = '&lt;button&gt;Fancy&lt;/button&gt;';
  const component = mount(<CodeBlock>{HTML}</CodeBlock>);

  expect(component.find('code').html()).toEqual(
    '<code class="code-block__code-content">&lt;button&gt;Fancy&lt;/button&gt;</code>'
  );
});

it('renders pre-escaped content with initial blank line', () => {
  const HTML = `
authors:
  - christoph@atelierdisko.de
  - mariuswilms@mailbox.org

description: &gt;
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.
`;
  const component = mount(<CodeBlock>{HTML}</CodeBlock>);

  const expected = `authors:
  - christoph@atelierdisko.de
  - mariuswilms@mailbox.org

description: &gt;
  This is a very very very fancy component. Lorem ipsum dolor sit amet,
  sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
  magna aliquyam erat, sed diam voluptua.
`;
  expect(component.find('code').html()).toEqual(`<code class="code-block__code-content">${expected}</code>`);
});
