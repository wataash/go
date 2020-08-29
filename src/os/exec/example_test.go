// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func Example_wataash_os_exec_filter_todo_merge() {
	cmd := exec.Command("sed", "s/foo/bar/g")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer func() {
			stdin.Close()
			wg.Done()
		}()
		if _, err := io.WriteString(stdin, "foo foo"); err != nil {
			log.Fatal(err)
		}
	}()

	var b []byte
	go func() {
		defer func() {
			stdout.Close()
			wg.Done()
		}()
		if b, err = ioutil.ReadAll(stdout); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", b)

	// Output: bar bar
}

var complexCmd = "echo 1 && sleep 1 && " +
	"echo 2 >&2 && sleep 1 && " +
	"echo 3  >&2 && sleep 1 && " +
	"echo 4"

func Example_wataash_osExec_Stdout() {
	cmd := exec.Command("sh", "-c", complexCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// errStart := cmd.Start()
	// _ = errStart
	errWait := cmd.Wait()
	_ = errWait

	// Output:
}

// https://tkuchiki.hatenablog.com/entry/2014/11/10/123447
func Example_wataash_osExec_Pipe() {
	cmd := exec.Command("sh", "-c", complexCmd)

	// stdout := &bytes.Buffer{} // non-blocking in scanner.Scan
	// p := make([]byte, 10)
	// n, err := stdout.Read(p) // io.EOF
	// _ = n

	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err != nil {
		panic(err)
	}

	// without wg, cmd.Wait() ends process, discarding buffer, scanner.Text()
	// ends due to io.EOF
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			out := scanner.Text()
			n, err := fmt.Fprintf(os.Stderr, "out: %s\n", out)
			_, _ = n, err
		}
		err := scanner.Err()
		_ = err
		wg.Done()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			out := scanner.Text()
			n, err := fmt.Fprintf(os.Stderr, "err: %s\n", out)
			_, _ = n, err
		}
		err := scanner.Err()
		_ = err
		wg.Done()
	}()

	// errStart := cmd.Start()
	// _ = errStart
	// wg.Wait()             // needed
	// errWait := cmd.Wait() // otherwise cmd.Wait discards command's output
	// _ = errWait

	// Output:
}

func Example_wataash_osExec_Tee() {
	cmd := exec.Command("sh", "-c", complexCmd)
	var bufOut, bufErr bytes.Buffer
	wg := &sync.WaitGroup{}
	wg.Add(2)
	pipeOut, _ := cmd.StdoutPipe()
	pipeErr, _ := cmd.StderrPipe()

	// _ = cmd.Start()

	go func() {
		teeOut := io.TeeReader(pipeOut, &bufOut)
		scanner := bufio.NewScanner(teeOut)
		for scanner.Scan() {
			fmt.Printf("out: %s\n", scanner.Text())
		}
		wg.Done()
	}()
	go func() {
		teeErr := io.TeeReader(pipeErr, &bufErr)
		scanner := bufio.NewScanner(teeErr)
		for scanner.Scan() {
			_, _ = fmt.Fprintf(os.Stderr, "err: %s\n", scanner.Text())
		}
		wg.Done()
	}()

	var err error

	// wg.Wait()
	// err = cmd.Wait()

	if errExit, ok := err.(*exec.ExitError); ok {
		rc := errExit.ExitCode()
		_ = rc
	} // else, err in exec

	s := bufOut.String() // 1 4
	s = bufErr.String() // 2 3
	_ = s

	// Output:
}

// http://www.albertoleal.me/posts/golang-pipes.html
func Example_wataash_osExec_PipeExtraFiles() {
	var err error
	var n int
	b := make([]byte, 128)

	// cmd := exec.Command("sh", "-c", "sleep 9999")
	cmd := exec.Command("sh", "-c", "echo foo >&3")

	pipeOut, err := cmd.StdoutPipe()
	pipeErr, err := cmd.StderrPipe()

	pipeR, pipeW, _ := os.Pipe()
	cmd.ExtraFiles = []*os.File{
		pipeW,
	}

	// [1]
	// Run -> `sleep 9999` doesn't have fd 3 pipe
	// expect stderr: "sh: 1: 3: Bad file descriptor"
	//   but not... due to no tty?
	// err = pipeW.Close()

	err = cmd.Run()

	// [2] should be here!
	err = pipeW.Close()

	// b, err = ioutil.ReadAll(pipeOut)
	// fmt.Fprintf(os.Stderr, "%s\n", b)

	n, err = pipeOut.Read(b)
	fmt.Fprintf(os.Stderr, "out: %s\n", b[:n])
	// fmt.Fprintf(os.Stderr, "out: %v\n", b[:n])
	n, err = pipeErr.Read(b)
	fmt.Fprintf(os.Stderr, "err: %s\n", b[:n])

	//                        [1]       [2]                [no close]
	n, err = pipeR.Read(b) // (0, EOF)  (4 ("foo\n", nil)  (4 ("foo\n", nil)
	fmt.Fprintf(os.Stderr, "pipeR: %s\n", b[:n])
	n, err = pipeR.Read(b) // (0, EOF)  (0, EOF)           block
	fmt.Fprintf(os.Stderr, "pipeR: %s\n", b[:n])

	err = pipeR.Close()

	_ = err
}

// http://www.albertoleal.me/posts/golang-pipes.html
func Example_wataash_osExec_PipeProcfs() {
	b, err := ioutil.ReadFile("/proc/self/stat")

	fd9 := os.NewFile(9, "/proc/self/fd/9")
	n, err := fd9.Write([]byte("hello")) // EBADF
	err = fd9.Close()                    // EBADF

	_ = b
	_ = err
	_ = n
}

// http://www.albertoleal.me/posts/golang-pipes.html
func Example_wataash_osExec_MkFifo() {
	var wg sync.WaitGroup
	tmpDir, err := ioutil.TempDir("", "named-pipes")
	_ = err

	// Create named pipe
	namedPipe := filepath.Join(tmpDir, "stdout")
	err = syscall.Mkfifo(namedPipe, 0600)

	wg.Add(1)
	go func() {
		_, err = fmt.Fprintln(os.Stderr, "go")
		cmd := exec.Command("sh", "-c", fmt.Sprintf("echo foo to %s ; echo foo > %s ; echo done", namedPipe, namedPipe))
		// Just to forward the stdout
		cmd.Stdout = os.Stdout
		_, err = fmt.Fprintln(os.Stderr, "go run")
		err = cmd.Run() // FIXME: deadlock
		_, err = fmt.Fprintln(os.Stderr, "go done")
		wg.Done()
	}()

	// Open named pipe for reading
	fmt.Println("Opening named pipe for reading")
	wg.Wait() // FIXME: deadlock
	stdout, err := os.OpenFile(namedPipe, os.O_RDONLY, 0600)
	// stdout, err := os.OpenFile(namedPipe, os.O_RDONLY|syscall.O_NONBLOCK, 0600)
	fmt.Println("Reading")

	var buff bytes.Buffer
	fmt.Println("Waiting for someone to write something")
	io.Copy(&buff, stdout)
	stdout.Close()
	fmt.Printf("Data: %s\n", buff.String())
}

func ExampleLookPath() {
	path, err := exec.LookPath("fortune")
	if err != nil {
		// log.Fatal("installing fortune is in your future")
		log.Print("installing fortune is in your future")
	}
	fmt.Printf("fortune is available at %s\n", path)
	// Output: fortune is available at
}

func ExampleCommand() {
	cmd := exec.Command("tr", "a-z", "A-Z")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("in all caps: %q\n", out.String())
	// Output: in all caps: "SOME INPUT"
}

func ExampleCommand_environment() {
	cmd := exec.Command("prog")
	cmd.Env = append(os.Environ(),
		"FOO=duplicate_value", // ignored
		"FOO=actual_value",    // this value is used
	)
	if err := cmd.Run(); err != nil {
		// log.Fatal(err)
		log.Print(err)
	}
	// Output:
}

func ExampleCmd_Output() {
	out, err := exec.Command("date").Output()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("The date is %s\n", out)
	out = []byte("2019年  2月  4日 月曜日 08:29:02 JST")
	fmt.Printf("The date is %s\n", out)
	// Output: The date is 2019年  2月  4日 月曜日 08:29:02 JST
}

func ExampleCmd_Run() {
	cmd := exec.Command("sleep", "1")
	log.Printf("Running command and waiting for it to finish...")
	err := cmd.Run()
	log.Printf("Command finished with error: %v", err)
	// Output:
}

func ExampleCmd_Start() {
	// cmd := exec.Command("sleep", "5")
	cmd := exec.Command("sleep", "1")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
	// Output:
}

func ExampleCmd_StdoutPipe() {
	cmd := exec.Command("echo", "-n", `{"Name": "Bob", "Age": 32}`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	var person struct {
		Name string
		Age  int
	}
	if err := json.NewDecoder(stdout).Decode(&person); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %d years old\n", person.Name, person.Age)
	// Output: Bob is 32 years old
}

func ExampleCmd_StdinPipe() {
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
	// Output: values written to stdin are passed to cmd's standard input
}

func ExampleCmd_StderrPipe() {
	cmd := exec.Command("sh", "-c", "echo stdout; echo 1>&2 stderr")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := io.ReadAll(stderr)
	fmt.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	// Output: stderr
}

func ExampleCmd_CombinedOutput() {
	cmd := exec.Command("sh", "-c", "echo stdout; echo 1>&2 stderr")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", stdoutStderr)
	// Output:
	// stdout
	// stderr
}

func ExampleCommandContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "sleep", "5").Run(); err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
	}
	// Output:
}
