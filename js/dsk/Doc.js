/**
 * Copyright 2017 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

export default class DocTransformer {

  // Constructor.
  //
  // Available options are:
  // - `modifier`, a function that receives the current node and prods and must
  //   return an object with the possibly modified `node` and `props` properties.
  //
  // - `noTransform`, a function that receives the type and props and
  //   must return a DOM node. The function is called whenever there is no
  //   transformation found for a node.
  //
  //   The following code turns a node into a React element, which is necessary
  //   because the parent might be a React element, which can’t have regular DOM
  //   nodes in props.children.
  //
  //   ```
  //   new Doc(..., ..., {
  //     noTransform: function(type, props) {
  //       return React.createElement(type, props, props.children);
  //     }
  //   });
  //   ```
  constructor(html, transforms, options = {}) {
    this.html = html;

    // Turn all the keys of our transforms into lowercase, because that’s how
    // HTML is parsed.
    this.transforms = {};

    let key, keys = Object.keys(transforms);
    let n = keys.length;
    while (n--) {
      key = keys[n];
      this.transforms[key.toLowerCase()] = transforms[key];
    }

    this.options = options;
  }

  compile() {
    // Use the browsers machinery to parse HTML and allow us to iterate
    // over it easily. Later child nodes are unwrapped from body again.
    let body = document.createElement('body');
    body.innerHTML = this.html;

    body.innerHTML = this.orphans(body);
    this.clean(body);

    let children = [];
    body.childNodes.forEach((c) => {
      let t = this.transform(c);
      if (t) {
        children.push(t);
      }
    });
    return children;
  }

  // Sometimes children are unnecessarily wrapped into another element.
  // If we find such elements, we unwrap them from their parent.
  //
  // This modifies the tree above the current node. Thus breaks our
  // tree walk and must be executed in a separate step.
  orphans(root) {
    if (!this.transforms.orphans) {
      return root.innerHTML;
    }

    let orphans = root.querySelectorAll(this.transforms.orphans.join(','));

    orphans.forEach((c) => {
      console.log(`Unwrapping ${c} from ${c.parentNode}`);

      // > If the given child is a reference to an existing node in the
      //   document, insertBefore() moves it from its current position to the new
      //   position (there is no requirement to remove the node from its parent
      //   node before appending it to some other node).
      //   - https://developer.mozilla.org/en-US/docs/Web/API/Node/insertBefore
      c.parentNode.parentNode.insertBefore(c, c.parentNode);
    });
    return root.innerHTML;
  }

  // If node is a <div>, we infer its type from its first attribute. This
  // means a Markdown doc can contain something like <div FullColorPlane> and
  // the transform will treat it as <FullColorPlane>.
  //
  // Please note, that the resulting element will have a lowercased tag name.
  isCustomElementCompat(node) {
    if (node.tagName !== 'DIV') {
      return false;
    }
    if (!node.attributes[0]) {
      return false;
    }
    if (!this.transforms[node.attributes[0].name.toLowerCase()]) {
      console.log(`Unknown custom element ${node.attributes[0].name}`);
      return false;
    }
    return true;
  }

  // Removes any elements, that may have become empty due to other
  // processing steps.
  clean(root) {
    root.querySelectorAll('p:empty').forEach((el) => {
      el.remove();
    });
  }

  // Replaces a given DOM node by applying a "transform".
  transform(node) {
    // Ignore nodes that only contain whitespace.
    if (node.nodeType === Node.TEXT_NODE && !node.nodeValue.trim()) {
      // Allow single spaces, for example between inline-elements
      if (node.nodeValue !== ' ') {
        return null;
      }
    }
    if (!node.tagName) {
      return node.textContent;
    }
    let type = node.tagName.toLowerCase();
    let props = { children: [] };

    // Where at the attributes should we begin parsing into props.
    let startProps = 0;

    if (this.isCustomElementCompat(node)) {
      type = node.attributes[0].name;
      startProps = 1;
    }

    let apply = this.transforms[type];

    // Turn node attributes into props object.
    for (let i = startProps; i < node.attributes.length; i++) {
      props[node.attributes[i].name] = node.attributes[i].value;
    }

    node.childNodes.forEach((c) => {
      let t = this.transform(c);
      if (t) {
        props.children.push(t);
      }
    });

    // If the node has no children, insert the text content as children.
    if (!props.children.length) {
      props.children = node.textContent || undefined;
    }

    // If the node has no key, we create a random one.
    if (!props.key) {
      props.key = Math.random();
    }

    if (this.options.modifier) {
      let modification = this.options.modifier(node, props);
      node = modification.node;
      props = modification.props;
    }

    // If there is no transform for the node, ignore it but still do include the
    // node in the final result.
    if (!apply) {
      if (this.options.noTransform) {
        return this.options.noTransform(type, props);
      }
      console.log(`No transform to apply to ${type}`);
      // FIXME: Children may be transformable; don't stop here.
    }

    return apply(props);
  }
}

export function transform(html, transforms, options = {}) {
  return (new DocTransformer(html, transforms, options)).compile();
}
