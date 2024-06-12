// Write command output to file.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var arg0 = filepath.Base(os.Args[0])

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:", arg0, "<file> <command> [<arg>]...")
	if w == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

func fatal(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	args := os.Args[1:]
	nargs := len(args)
	if nargs > 0 && args[0] == "-h" {
		usage(os.Stdout)
	}
	if nargs < 2 {
		usage(os.Stderr)
	}
	output, err := os.Create(args[0])
	if err != nil {
		fatal(err)
	}
	defer output.Close()
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdout = output
	if err = cmd.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			fatal(eerr.Stderr)
		} else {
			fatal(err)
		}
	}
}
