// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides a flag.FlagSet wrapper that doesn't output during
// Parse and instead sets an associate `help` flag.  It also provides a
// fmt.Formatter interface to PrintDefaults.
package flag

import (
	"flag"
	"fmt"
	"io"
)

const helpmsg = "Print options."

type Flag = flag.Flag
type FlagSet struct{ *flag.FlagSet }

var CommandLine = FlagSet{flag.CommandLine}

func New() (fs FlagSet, help *bool) {
	fs.FlagSet = flag.NewFlagSet("", 0)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	help = fs.Bool("help", false, helpmsg)
	fs.BoolVar(help, "h", *help, helpmsg)
	return
}

func (fs FlagSet) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w)
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(io.Discard)
}
