// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/internal/liner"
)

func main() {
	goes.Plot(builtinutils.New()...)
	fmt.Println("Type EOF to exit...")
	l := liner.New()
	prompt := filepath.Base(os.Args[0]) + "> "
	for {
		s, err := l.GetLine(prompt)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			break
		}
		fmt.Printf("input: %q\n", s)
	}
}
