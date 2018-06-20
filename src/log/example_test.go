// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func ExampleWataashLog() {
	logger := log.New(os.Stderr, "-- ", log.LstdFlags)
	// simplest:
	// logger := log.New(os.Stderr, "", 0)

	// same: -- 2019/02/03 13:18:40 foo
	logger.Println("foo")
	// logger.Println("foo\n") // error: redundant newline
	logger.Print("foo") // appends \n. see (*Logger).Output()
	logger.Print("foo\n")

	// Output:
}

// https://blog.satotaichi.info/logging-frameworks-for-go/
func ExampleWataashLogUseGlobalLogger() {
	log.SetPrefix("-- ")

	log.SetFlags(log.Ldate | log.Ltime)                     // -- 2018/08/02 00:16:07 ++
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)    // -- 2018/08/02 00:16:35 main.go:18: ++
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds) // -- 2018/08/02 00:16:56.350494 ++
	log.SetFlags(log.Lmicroseconds)                         // -- 00:17:09.831360 ++
	log.SetFlags(log.Ltime | log.LUTC)                      // -- 15:17:53 ++

	log.Println("++")
	// Output:
}

// https://blog.satotaichi.info/logging-frameworks-for-go/
func ExampleWataashLogUseLogger() {
	logger := log.New(os.Stderr, "**** ", log.LstdFlags|log.Lshortfile)
	// **** 2019/02/03 13:10:23 example_test.go:36: Hey!! logging
	logger.Printf("Hey!! %s", "logging")
	// logger.Fatalf("Hey!! %s", "logging")

	// Output:
}

func ExampleLogger() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	logger.Print("Hello, log file!")

	fmt.Print(&buf)
	// Output:
	// logger: example_test.go:19: Hello, log file!
}

func ExampleLogger_Output() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "INFO: ", log.Lshortfile)

		infof = func(info string) {
			logger.Output(2, info)
		}
	)

	infof("Hello world")

	fmt.Print(&buf)
	// Output:
	// INFO: example_test.go:36: Hello world
}
