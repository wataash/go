// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

// goland remote debug tesgo:
//   dlv debug --headless --listen=:2345 --api-version=2 github.com/wataash/tesgo
// goland remote test os:
//   dlv test --headless --listen=:2345 --api-version=2 os -- -test.run '^Example'
//
// TODO: without goland: dlv {debug, test}
//
// EBADF0: &os.PathError{"read", "/dev/stdin", syscall.EBADF}
// EBADF1: &os.PathError{"read", "/dev/stdout", syscall.EBADF}
// EBADF2: &os.PathError{"read", "/dev/stderr", syscall.EBADF}
// EBADF|1: &os.PathError{"read", "|1", syscall.EBADF}
//
//                 goland debug tesgo  goland test os  goland remote debug tesgo  goland remote test os
// os.Stdin.name:  "/dev/stdin"        "/dev/stdin"    "/dev/stdin"               "/dev/stdin"
// os.Stdout.name: "/dev/stdout"       "|1"            "/dev/stdout"              "/dev/stdout"
// os.Stderr.name: "/dev/stderr"       "/dev/detout"   "/dev/stderr"              "/dev/stderr"
// in.Write        0, EBADF0           0, EBADF0       15, nil                    15, nil
// out.Write       16, nil             16(piped), nil  16, nil                    16(piped), nil
// err.Write       16, nil             16, nil         16, nil                    16, nil
// in.Read         io.EOF              io.EOF          n(block), nil              n(block), nil
// out.Read        0, EBADF1           0, EBADF|1      n(block), nil              0, EBADF|1
// err.Read        0, EBADF2           0, EBADF2       n(block), nil              n(block), nil
func ExampleWataAshOs() {
	var n int
	var err error

	n, err = os.Stdin.Write([]byte("os.Stdin.Write\n"))
	n, err = os.Stdout.Write([]byte("os.Stdout.Write\n"))
	n, err = os.Stderr.Write([]byte("os.Stderr.Write\n"))

	b := make([]byte, 100)
	n, err = os.Stdin.Read(b)
	n, err = os.Stdout.Read(b)
	n, err = os.Stderr.Read(b)

	_, _ = n, err
	// Output:
}

func ExampleOpenFile() {
	f, err := os.OpenFile("notes.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleOpenFile_append() {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte("appended some data\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleChmod() {
	if err := os.Chmod("some-filename", 0644); err != nil {
		log.Fatal(err)
	}
}

func ExampleChtimes() {
	mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
	atime := time.Date(2007, time.March, 2, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes("some-filename", atime, mtime); err != nil {
		log.Fatal(err)
	}
}

func ExampleFileMode() {
	fi, err := os.Lstat("some-filename")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("permissions: %#o\n", fi.Mode().Perm()) // 0400, 0777, etc.
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		fmt.Println("regular file")
	case mode.IsDir():
		fmt.Println("directory")
	case mode&fs.ModeSymlink != 0:
		fmt.Println("symbolic link")
	case mode&fs.ModeNamedPipe != 0:
		fmt.Println("named pipe")
	}
}

func ExampleIsNotExist() {
	filename := "a-nonexistent-file"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("file does not exist")
	}
	// Output:
	// file does not exist
}

func ExampleExpand() {
	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "DAY_PART":
			return "morning"
		case "NAME":
			return "Gopher"
		}

		return ""
	}

	fmt.Println(os.Expand("Good ${DAY_PART}, $NAME!", mapper))

	// Output:
	// Good morning, Gopher!
}

func ExampleExpandEnv() {
	os.Setenv("NAME", "gopher")
	os.Setenv("BURROW", "/usr/gopher")

	fmt.Println(os.ExpandEnv("$NAME lives in ${BURROW}."))

	// Output:
	// gopher lives in /usr/gopher.
}

func ExampleLookupEnv() {
	show := func(key string) {
		val, ok := os.LookupEnv(key)
		if !ok {
			fmt.Printf("%s not set\n", key)
		} else {
			fmt.Printf("%s=%s\n", key, val)
		}
	}

	os.Setenv("SOME_KEY", "value")
	os.Setenv("EMPTY_KEY", "")

	show("SOME_KEY")
	show("EMPTY_KEY")
	show("MISSING_KEY")

	// Output:
	// SOME_KEY=value
	// EMPTY_KEY=
	// MISSING_KEY not set
}

func ExampleGetenv() {
	os.Setenv("NAME", "gopher")
	os.Setenv("BURROW", "/usr/gopher")

	fmt.Printf("%s lives in %s.\n", os.Getenv("NAME"), os.Getenv("BURROW"))

	// Output:
	// gopher lives in /usr/gopher.
}

func ExampleUnsetenv() {
	os.Setenv("TMPDIR", "/my/tmp")
	defer os.Unsetenv("TMPDIR")
}

func ExampleReadDir() {
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func ExampleMkdirTemp() {
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	file := filepath.Join(dir, "tmpfile")
	if err := os.WriteFile(file, []byte("content"), 0666); err != nil {
		log.Fatal(err)
	}
}

func ExampleMkdirTemp_suffix() {
	logsDir, err := os.MkdirTemp("", "*-logs")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(logsDir) // clean up

	// Logs can be cleaned out earlier if needed by searching
	// for all directories whose suffix ends in *-logs.
	globPattern := filepath.Join(os.TempDir(), "*-logs")
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		log.Fatalf("Failed to match %q: %v", globPattern, err)
	}

	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			log.Printf("Failed to remove %q: %v", match, err)
		}
	}
}

func ExampleCreateTemp() {
	f, err := os.CreateTemp("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	if _, err := f.Write([]byte("content")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleCreateTemp_suffix() {
	f, err := os.CreateTemp("", "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	if _, err := f.Write([]byte("content")); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleReadFile() {
	data, err := os.ReadFile("testdata/hello")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)

	// Output:
	// Hello, Gophers!
}

func ExampleWriteFile() {
	err := os.WriteFile("testdata/hello", []byte("Hello, Gophers!"), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
