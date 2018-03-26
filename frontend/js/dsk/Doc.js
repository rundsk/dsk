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

    this.orphans(body);
    // this.customElements(body);
    this.clean(body);
    this.tranform(body);

    return body.childNodes;
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
      c.parentNode.insertBefore(c, c.parentNode);
    });
  }

  // If node is a <div>, we infer its type from its first attribute. This
  // means a Markdown doc can contain something like <div FullColorPlane> and
  // the transform will treat it as <FullColorPlane>.
  //
  // Please note, that the resulting element will have a lowercased tag name.
  //  customElements(root) {
  //    let isCustomElement = (
  //      node.tagName === 'DIV'
  //      && node.attributes[0]
  //      && this.transforms[node.atttributes[0].name.toLowerCase()]
  //    );
  //    if (!isCustomElement) {
  //      return node;
  //    }
  //    // > When called on an HTML document, createElement() converts
  //    //   tagName to lower case before creating the element.
  //    let custom = document.createElement(node.attributes[0]);
  //
  //    // Copy over all children.
  //    node.childNodes.forEach((c) => {
  //      custom.appendChild(c);
  //    });
  //
  //    // Skip the first attribute, we used it for the element name above.
  //    for (let i = 1; i < node.attributes.length; i++) {
  //      custom.setAttribute(node.attributes[i].name, node.attributes[i].value);
  //    }
  //    node.parentNode.replaceChild(custom, node);
  //  }

  // Removes any elements, that may have become empty due to other
  // processing steps.
  clean(root) {
    root.querySelectorAll('*:empty').remove();
  }

  // Replaces a given DOM node by applying a "transform".
  transform(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      return;
    }
    let apply = this.transforms[node.tagName.toLowerCase()];

    // If there is no transform for the node, ignore it but still do
    // include the node in the final result.
    if (!apply) {
      console.log(`No transform to apply to ${node}`);
      // Children may be transformable; don't stop here.

      node.childNodes.forEach((c) => {
        node.replaceChild(this.transform(c), c);
      });
      return;
    }
    let props = {};

    // Turn node attributes into props object.
    for (let i = 0; i < node.attributes.length; i++) {
      props[node.attributes[i].name] = node.attributes[i].value;
    }

    // Descend first.
    node.childNodes.forEach((c) => {
      node.replaceChild(this.transform(c), c);
    });
    node.parentNode.replaceChild(apply(node, props), node);
  }
}

export default function transform(html, transforms) {
  return (new DocTransformer(html, transforms)).compile();
}

