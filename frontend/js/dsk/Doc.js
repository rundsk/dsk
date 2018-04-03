/**
 * Copyright 2017 Atelier Disko. All rights reserved. This source
 * code is distributed under the terms of the BSD 3-Clause License.
 */

class DocTransformer {
  constructor(html, transforms) {
    this.html = html;
    this.transforms = transforms;
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
    let props = { key: Math.random() };
    let children = [];

    // Where at the attributes should we begin parsing into props.
    let startProps = 0;

    if (this.isCustomElementCompat(node)) {
      type = node.attributes[0].name;
      startProps = 1;
    }

    let apply = this.transforms[type];

    // If there is no transform for the node, ignore it but still do
    // include the node in the final result.
    if (!apply) {
      console.log(`No transform to apply to ${type}`);
      // FIXME: Children may be transformable; don't stop here.
      return null;
    }

    // Turn node attributes into props object.
    for (let i = startProps; i < node.attributes.length; i++) {
      props[node.attributes[i].name] = node.attributes[i].value;
    }

    node.childNodes.forEach((c) => {
      let t = this.transform(c);
      if (t) {
        children.push(t);
      }
    });

    // If the node has no children, insert the text content as children
    if (!children.length) {
      children = node.textContent;
    }
    return apply(node, children, props);
  }
}

export default function transform(html, transforms) {
  return (new DocTransformer(html, transforms)).compile();
}
