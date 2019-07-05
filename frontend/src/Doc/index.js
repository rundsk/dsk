import React, { useEffect } from 'react';
import ReactDOM from 'react-dom';
import { transform } from '@atelierdisko/dsk';
import { BaseLink, withRoute } from 'react-router5';
import { slugify } from '../utils';
import { copyTextToClipboard } from '../utils';

import Banner from '../Banner'

import './Doc.css';
import ComponentDemo from '../ComponentDemo';
import TypographySpecimen from '../TypographySpecimen';
import CodeBlock from '../CodeBlock';
import FigmaEmbed from '../FigmaEmbed';
import DoDont, { Do, Dont } from '../DoDont';
import ColorGroup from '../ColorGroup';
import ColorCard from '../ColorCard';

// import MDX from '@mdx-js/runtime';

function Doc(props) {
  const ref = React.createRef();

  useEffect(() => {
    // Clean up from the previous doc
    if (ref.current) {
      ReactDOM.unmountComponentAtNode(ref.current);
    }

    renderComponents();
    setupImages();
    replaceInternalLinks();
    replaceHeadings();
    makeCodeCopyable();
  }, [props.title]);

  // Find all images whose src includes '@2x' and set
  // their height and width so they are displayed @2x
  function setupImages() {
    if (ref.current) {
      let node = ref.current;

      // Find all images inside p’s and unwrap them
      let orphans = ref.current.querySelectorAll("p > img, p > video");
      orphans.forEach((o) => {
        // > If the given child is a reference to an existing node in the
        //   document, insertBefore() moves it from its current position to the new
        //   position (there is no requirement to remove the node from its parent
        //   node before appending it to some other node).
        //   - https://developer.mozilla.org/en-US/docs/Web/API/Node/insertBefore
        o.parentNode.parentNode.insertBefore(o, o.parentNode);
      });

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
      let imgs = node.querySelectorAll("img");
      imgs.forEach(img => {
        let src = img.getAttribute("src");

        if (src.includes("@2x")) {
          let width = img.getAttribute("width");
          let height = img.getAttribute("height");

          if (!width || !height) { return; }

          img.style.maxWidth = `${width / 2}px`;
          img.style.maxHeight = `${height / 2}px`;
        }
      });
    }
  }

  // Finds all code section that have been tagged Component,
  // parse them and mount react components
  function renderComponents() {
    if (!props.mountComponents) {
      return;
    }

    if (!ref.current) {
      return;
    }
    let doc = ref.current;

    // Find retina images and set them to display at half
    // their size. The information about their width and height
    // is added by the dsk back-end.

    const transforms = {
      Banner: props => { return <Banner {...props} />},
      Warning: props => <Banner type="warning" {...props} />,
      Playground: props => <ComponentDemo {...props} />,
      TypographySpecimen: props => <TypographySpecimen {...props} />,
      FigmaEmbed: props => <FigmaEmbed {...props} />,
      CodeBlock: props => <CodeBlock {...props} />,
      DoDontGroup: props => <DoDont {...props} />,
      Do: props => <Do {...props} />,
      Dont: props => <Dont {...props} />,
      ColorGroup: props => <ColorGroup {...props} />,
      Color: props => <ColorCard {...props} />
    };

    let transformedContent = transform(doc.innerHTML, transforms, {
      noTransform: (type, props) => {
        // This gets called on HTML elements that do not need
        // to be transformed to special React components.
        // There are differences between the attributes of
        // HTML elements and React that we have to take care
        // of: https://reactjs.org/docs/dom-elements.html#differences-in-attributes
        props.className = props.class;
        delete(props.class);

        return React.createElement(type, props, props.children);
      }
    });
    ReactDOM.render(transformedContent, doc);
  }

  // Replace links to internal node with links from the router
  function replaceInternalLinks() {
    if (ref.current) {
      let doc = ref.current;

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
      let links = doc.querySelectorAll("a[data-node]");
      if (links.length === 0) { return; }

      links.forEach(l => {
        let content = l.textContent;
        let href = l.getAttribute("href");

        let hash = href.split("?t=")[1] || undefined;
        let path = l.getAttribute("data-node");
        let routerLink = <BaseLink router={props.router} routeName='node' routeParams={{ node: path, t: hash }}>{content}</BaseLink>

        let newNode = document.createElement("span");
        l.parentNode.replaceChild(newNode, l);

        ReactDOM.render(routerLink, newNode, () => {
          // After rendering we want to unwrap the children we just
          // created from the a around them so they are placed
          // directly in the flow of the document
          if (newNode.childNodes.length > 0) {
            newNode.replaceWith(...newNode.childNodes);
          }
        });
      });
    }
  }

  // Add a jumplink to each heading
  function replaceHeadings() {
    if (ref.current) {
      let doc = ref.current;

      let heading = doc.querySelectorAll("h1, h2, h3, h4, h5");
      if (heading.length === 0) { return; }

      heading.forEach(h => {
        if (h.classList.contains("heading")) { return; }
        h.classList.add("heading");

        if (h.getAttribute("nojumplink") === "true") { return; }

        let link = document.createElement("div");
        link.textContent = "§";
        link.classList.add("heading__jumplink");

        let id = slugify(h.textContent);
        h.id = id;
        // We can’t use just id, because it doesn’t work for number-only headings
        h.setAttribute("heading-id", id);

        link.onclick = () => {
          let currentRouterState = props.router.getState();
          let currentNode = currentRouterState.params.node || "";
          let t = slugify(props.title) + "§" + id;

          props.router.navigate("node", { ...currentRouterState.params, node: currentNode, t: t }, { replace: true });
        };

        h.prepend(link);
      });
    }
  }

  // Add a 'copy' link to every code section
  function makeCodeCopyable() {
    if (ref.current) {
      let doc = ref.current;

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
      let code = doc.querySelectorAll("pre code");
      if (code.length === 0) { return; }

      code.forEach(c => {
        if (c.classList.contains("code--is-copyable")) { return; }
        // The Code Block component has its own copy link
        if (c.classList.contains("code-block__code-content")) { return; }
        c.classList.add("code--is-copyable");

        if (c.getAttribute("dontcopy") === "true") { return; }

        let codeWrapper = document.createElement("div");
        codeWrapper.classList.add("code");

        let copyLink = document.createElement("div");
        copyLink.textContent = "Copy";
        copyLink.classList.add("code__copyLink");

        copyLink.onclick = () => {
          copyTextToClipboard(c.textContent);
          copyLink.textContent = "Copied!";

          setTimeout(() => {
            copyLink.textContent = "Copy";
          }, 2000);
        };

        c.parentNode.replaceWith(codeWrapper);
        codeWrapper.prepend(c.parentNode);
        codeWrapper.prepend(copyLink);

        // Insert it in the pre, not in the code
        //doc.insertBefore(copyLink, c.parentNode);
      });
    }
  }

  // This does nor work currently, because of issues with MDX-runtime and
  // create-react-app: https://spectrum.chat/mdx/general/unable-to-run-mdx-js-runtime-example-using-webpack~65dccf86-1226-4c9a-9af5-0ed9ef338ffb
  // renderComponentsWithMDX() {
  //   if (this.ref.current) {
  //     let node = this.ref.current;

  //     const components = {
  //       h1: props => <h1 style={{ color: 'tomato' }} {...props} />
  //     }

  //     const scope = {
  //       Warning: props => <Warning {...props} />
  //     }

  //     // Find retina images and set them to display at half
  //     // their size. The information about their width and height
  //     // is added by the dsk back-end.
  //     let code = node.querySelectorAll("code[class='language-Component']");
  //     code.forEach(c => {
  //       let content = c.textContent;
  //       // We want to replace the <pre> tag with our component
  //       let parentNode = c.parentNode.parentNode;

  //       ReactDOM.render(<MDX components={components} scope={scope}>{content}</MDX>, parentNode);
  //     });
  //   }
  // }

  // if there is a doc but it is empty we cannot risk calling dangerouslySetInnerHTML
  if (props.htmlContent === "") {
    return <div className="doc" ref={ref}>{props.children}</div>;
  }

  return (
    // We want to be able to fill docs straigt with HTML or with react
    // components, depending on which props are passed.
    <div className="doc" ref={ref} dangerouslySetInnerHTML={props.htmlContent && { __html: props.htmlContent }}>
      {props.children}
    </div>
  );
}

export default withRoute(Doc);
