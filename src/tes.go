package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "foo")
	fmt.Fprint(os.Stdout, "%T %p", 1, 1)
	// fmt.Fprintln(os.Stdout, "foo")
}
