// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httptrace_test

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"os"
)

func Example() {
	out := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = out }()

	// DNS Info: {Addrs:[{IP:93.184.216.34 Zone:} {IP:2606:2800:220:1:248:1893:25c8:1946 Zone:}] Err:<nil> Coalesced:false}
	// Got Conn: {Conn:0xc0000a4040 Reused:false WasIdle:false IdleTime:0s}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("Got Conn: %+v\n", connInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Info: %+v\n", dnsInfo)
		},
	}

	// GetConn: example.com:80
	// DNSStart: {Host:example.com}
	// DNSDone: {Addrs:[{IP:93.184.216.34 Zone:} {IP:2606:2800:220:1:248:1893:25c8:1946 Zone:}] Err:<nil> Coalesced:false}
	// ConnectStart: tcp 93.184.216.34:80
	// ConnectDone: tcp 93.184.216.34:80
	// GotConn: {Conn:0xc0000ac040 Reused:false WasIdle:false IdleTime:0s}
	// WroteHeaderField: Host, [example.com]
	// WroteHeaderField: User-Agent, [Go-http-client/1.1]
	// WroteHeaderField: Accept-Encoding, [gzip]
	// WroteHeaders:
	// WroteRequest: {Err:<nil>}
	// GotFirstResponseByte:
	trace = &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			fmt.Printf("GetConn: %+v\n", hostPort)
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("GotConn: %+v\n", connInfo)
		},
		PutIdleConn: func(err error) {
			// TODO: step into here
			fmt.Printf("PutIdleConn: %+v\n", err)
		},
		GotFirstResponseByte: func() {
			fmt.Printf("GotFirstResponseByte: \n")
		},
		Got100Continue: func() {
			// TODO: step into here
			fmt.Printf("Got100Continue: \n")
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			// TODO: step into here
			fmt.Printf("Got1xxResponse: %d, %+v\n", code, header)
			return nil
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			fmt.Printf("DNSStart: %+v\n", info)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			fmt.Printf("DNSDone: %+v\n", info)
		},
		ConnectStart: func(network, addr string) {
			fmt.Printf("ConnectStart: %s %s\n", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			fmt.Printf("ConnectDone: %s %s\n", network, addr)
		},
		TLSHandshakeStart: func() {
			// TODO: step into here
			fmt.Printf("TLSHandshakeStart: \n")
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			// TODO: step into here
			fmt.Printf("TLSHandshakeDone: \n")
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Printf("WroteHeaderField: %s, %s\n", key, value)
		},
		WroteHeaders: func() {
			fmt.Printf("WroteHeaders: \n")
		},
		Wait100Continue: func() {
			fmt.Printf("Wait100Continue: \n")
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			fmt.Printf("WroteRequest: %+v\n", info)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// tmp1 := req.Context()
	// tmp2 := httptrace.WithClientTrace(tmp1, trace)
	// req = req.WithContext(tmp2)

	_, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Fatal(err)
	}

	// DNS Info: {Addrs:[{IP:93.184.216.34 Zone:} {IP:2606:2800:220:1:248:1893:25c8:1946 Zone:}] Err:<nil> Coalesced:false}
	// Got Conn: {Conn:0xc0000a4040 Reused:false WasIdle:false IdleTime:0s}

	// Output:
}
