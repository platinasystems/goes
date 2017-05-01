// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/flags"
	"github.com/platinasystems/go/goes/cmd/nldump"
)

var Exit = os.Exit
var Stderr io.Writer = os.Stderr

func main() {
	m := goes.New(flags.New()...)
	m.Plot(nldump.New())
	err := m.Main()
	if err != nil {
		fmt.Fprintln(Stderr, err)
		Exit(1)
	}
}
