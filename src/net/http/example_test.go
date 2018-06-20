// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func ExampleHijacker() {
	outOrig := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = outOrig }()
	// echo -e GET '/foo    HTTP/1.1\nHost: localhost:8080\n' |
	//   socat -v -v -x -t 10 - TCP4:localhost:8080,crnl
	// echo -e GET '/hijack HTTP/1.1\nHost: localhost:8080\n\nfoo\n' |
	//   socat -v -v -x -t 10 - TCP4:localhost:8080,crnl

	http.HandleFunc("/hijack", func(w http.ResponseWriter, r *http.Request) {
		// w: *http.response; is a http.Hijacker
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Don't forget to close the connection:
		defer conn.Close()
		bufrw.WriteString("Now we're speaking raw TCP. Say hi: ")
		bufrw.Flush()
		s, err := bufrw.ReadString('\n')
		if err != nil {
			log.Printf("error reading string: %v", err)
			return
		}
		fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
		bufrw.Flush()
	})

	http.ListenAndServe(":8080", nil)
	// // Output:
}

func ExampleGet() {
	outOrig := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = outOrig }()

	res, err := http.Get("http://www.google.com/robots.txt")
	if err != nil {
		log.Fatal(err)
	}
	robots, err := io.ReadAll(res.Body)
	res.Body.Close()

	// res.Body: io.ReadCloser | *http.gzipReader
	err = res.Body.Close()
	err = res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", robots)

	// // Output:
}

func ExampleFileServer() {
	outOrig := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = outOrig }()

	// http://localhost:8080/

	tmpDir := http.Dir("/usr/share/doc") // implements http.FileSystem
	_ = http.FileSystem.Open
	_ = tmpDir.Open
	tmpHandler := http.FileServer(tmpDir) // net/http.Handler | *net/http.fileHandler
	_ = http.Handler.ServeHTTP
	// _ = http.fileHandler.ServeHTTP // set breakpoint in it!
	log.Fatal(http.ListenAndServe(":8080", tmpHandler))

	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("/usr/share/doc"))))

	// // Output:
}

func ExampleFileServer_stripPrefix() {
	outOrig := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = outOrig }()

	// http://localhost:8080/tmpfiles/

	tmpHandler := http.FileServer(http.Dir("/tmp"))           // net/http.Handler | *net/http.fileHandler
	tmpHandler2 := http.StripPrefix("/tmpfiles/", tmpHandler) // net/http.Handler |  net/http.HandlerFunc
	// _ = http.fileHandler.ServeHTTP // breakpoint in it
	_ = http.HandlerFunc.ServeHTTP // breakpoint in it
	http.Handle("/tmpfiles/", tmpHandler2)

	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	// http.Handle("/tmpfiles/", http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))

	http.ListenAndServe(":8080", nil)
	// // Output:
}

func ExampleStripPrefix() {
	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	http.Handle("/tmpfiles/", http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
}

type apiHandler struct{}

func (handler apiHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "(apiHandler)ServeHTTP")
}

func ExampleServeMux_Handle() {
	// http://localhost:8080/
	// http://localhost:8080/foo/
	// http://localhost:8080/api/

	mux := http.NewServeMux()
	_ = (* http.ServeMux).ServeHTTP
	_ = (* http.ServeMux).Handle
	_ = (* http.ServeMux).HandleFunc
	_ = (* http.ServeMux).Handler
	mux.Handle("/api/", apiHandler{})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	_ = http.ListenAndServe(":8080", mux)
	// // Output:
}

// not tried below...

// HTTP Trailers are a set of key/value pairs like headers that come
// after the HTTP response, instead of before.
func ExampleResponseWriter_trailers() {
	// http http://localhost:8080/sendstrailers
	// curl -v --raw http://localhost:8080/sendstrailers
	// only curl shows the trailers

	mux := http.NewServeMux()
	mux.HandleFunc("/sendstrailers", func(w http.ResponseWriter, req *http.Request) {
		// Before any call to WriteHeader or Write, declare
		// the trailers you will set during the HTTP
		// response. These three headers are actually sent in
		// the trailer.
		w.Header().Set("Trailer", "AtEnd1, AtEnd2")
		w.Header().Add("Trailer", "AtEnd3")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)

		w.Header().Set("AtEnd1", "value 1")
		io.WriteString(w, "This HTTP response has both headers before this text and trailers at the end.\n")
		w.Header().Set("AtEnd2", "value 2")
		w.Header().Set("AtEnd3", "value 3") // These will appear as trailers.
	})

	_ = http.ListenAndServe(":8080", mux)
	// Output:
}

func ExampleServer_Shutdown() {
	outOrig := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = outOrig }()

	var srv http.Server

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	// // Output:
}

func ExampleListenAndServeTLS() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, TLS!\n")
	})

	// One can use generate_cert.go in crypto/tls to generate cert.pem and key.pem.
	log.Printf("About to listen on 8443. Go to https://127.0.0.1:8443/")
	err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
	log.Fatal(err)
	// // Output:
}

func ExampleListenAndServe() {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
	// // Output:
}

func ExampleHandleFunc() {
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #1!\n")
	}
	h2 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #2!\n")
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/endpoint", h2)

	log.Fatal(http.ListenAndServe(":8080", nil))
	// // Output:
}

func newPeopleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is the people handler.")
	})
	// // Output:
}

func ExampleNotFoundHandler() {
	mux := http.NewServeMux()

	// Create sample handler to returns 404
	mux.Handle("/resources", http.NotFoundHandler())

	// Create sample handler that returns 200
	mux.Handle("/resources/people/", newPeopleHandler())

	log.Fatal(http.ListenAndServe(":8080", mux))
	// // Output:
}
