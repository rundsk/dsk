System.register([], function (exports) {
  'use strict';
  return {
    execute: function () {

      /**
       * Copyright 2019 Atelier Disko. All rights reserved. This source
       * code is distributed under the terms of the BSD 3-Clause License.
       */
      function Button(props) {
        return React.createElement("button", null, "Hello DSK!");
      }
      exports('Button', Button);

    }
  };
});
