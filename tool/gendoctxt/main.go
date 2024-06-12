// Gendoctxt generates doc text files for a list of goes symbols.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var arg0 = filepath.Base(os.Args[0])

var Symbols = []string{
	"pkg/net/vpn.Registry",
	"pkg/net/vpn.Exchange",
	"pkg/net/vpn.Guest",
}

func fatal(args ...any) {
	if len(args) > 0 {
		if err, ok := args[0].(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, err.Stderr)
			os.Exit(1)
		}
	}
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	var dryrun bool
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h":
			fmt.Println("usage:", arg0, "[-dryrun]")
			return
		case "-dryrun":
			dryrun = true
		}
	}
	for _, sym := range Symbols {
		dot := strings.Index(sym, ".")
		pkg := sym[:dot]
		fn := fmt.Sprint(strings.ToLower(sym[dot+1:]), ".txt")
		pn := filepath.Join(pkg, fn)
		if dryrun {
			fmt.Println("go doc", sym, ">", pn)
		} else {
			output, err := os.Create(pn)
			if err != nil {
				fatal(err)
			}
			defer output.Close()
			cmd := exec.Command("go", "doc", sym)
			cmd.Stdout = output
			err = cmd.Run()
			if err != nil {
				fatal(err)
			}
		}
	}
}
