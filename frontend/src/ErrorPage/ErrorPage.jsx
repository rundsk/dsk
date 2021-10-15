/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import './ErrorPage.css';

function ErrorPage(props) {
  return (
    <div className="error-page">
      <div className="error-page__content">
        <h1 className="error-page__error-message">{props.children}</h1>
      </div>
    </div>
  );
}

export default ErrorPage;
