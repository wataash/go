// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type countHandler struct {
	mu sync.Mutex // guards n
	n  int
}

func (h *countHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.n++
	fmt.Fprintf(w, "count is %d\n", h.n)
}

func ExampleHandle() {
	out := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = out }()
	// http://localhost:8080/count

	http.Handle("/count", new(countHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))

	log.Fatal(nil)
	log.Fatal(errors.New("foo"))
	// // Output:
}
