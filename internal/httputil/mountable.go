// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import "net/http"

type Mountable interface {
	// HTTPMux returns a HTTP mux that can be mounted onto a root mux.
	HTTPMux() http.Handler
}
