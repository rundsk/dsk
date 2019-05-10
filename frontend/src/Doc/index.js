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
import ColorSpecimen from '../ColorSpecimen';
import CodeBlock from '../CodeBlock';
import FigmaEmbed from '../FigmaEmbed';
import DoDont, { Do, Dont } from '../DoDont';

// import MDX from '@mdx-js/runtime';

function Doc(props) {
  const ref = React.createRef();

  useEffect(() => {
    renderComponents();
    setupImages();
    replaceInternalLinks();
    replaceHeadings();
    makeCodeCopyable();
  });

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

  // FIXME: We need to unmount the components properly,
  // otherwise there might be trouble when switchting between
  // documents
  // Finds all code section that have been tagged Component,
  // parse them and mount react components
  function renderComponents() {
    if (ref.current) {
      let doc = ref.current;

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
      let code = doc.querySelectorAll("code[class='language-Component']");
      if (code.length === 0) { return; }

      const transforms = {
        Banner: props => { return <Banner {...props} />},
        Warning: props => <Banner type="warning" {...props} />,
        ComponentDemo: props => <ComponentDemo {...props} />,
        TypographySpecimen: props => <TypographySpecimen {...props} />,
        ColorSpecimen: props => <ColorSpecimen {...props} />,
        FigmaEmbed: props => <FigmaEmbed {...props} />,
        CodeBlock: props => <CodeBlock {...props} />,
        DoDont: props => <DoDont {...props} />,
        Do: props => <Do {...props} />,
        Dont: props => <Dont {...props} />,
      };

      code.forEach(c => {
        let content = c.textContent;

        // console.log(jsx`${content}`);

        // We want to replace the <pre> tag with our component
        let parentNode = c.parentNode;
        let newNode = document.createElement("div");
        parentNode.parentNode.replaceChild(newNode, parentNode);

        let transformedContent = transform(content, transforms, {
          noTransform: (type, props) => {
            return React.createElement(type, props, props.children);
          }
        });

        ReactDOM.render(transformedContent, newNode, () => {
          // After rendering we want to unwrap the children we just
          // created from the div around them so they are placed
          // directly in the flow of the document
          if (newNode.childNodes.length > 0) {
            newNode.replaceWith(...newNode.childNodes);
          }
        });
      });
    }
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

      // Find retina images and set them to display at half
      // their size. The information about their width and height
      // is added by the dsk back-end.
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
  if (props.content === "") {
    return <div className="doc" ref={ref}>{props.children}</div>;
  }

  return (
    <div className="doc" ref={ref} dangerouslySetInnerHTML={props.content && { __html: props.content }}>
      {props.children}
    </div>
  );
}

export default withRoute(Doc);
