// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package flags provides a flag.FlagSet wrapper that doesn't output during
// Parse, it only returns error or nil.  It also provides a fmt.Formatter
// interface to PrintDefaults.
package flags

import (
	"flag"
	"fmt"
	"io"
)

var ErrHelp = flag.ErrHelp

type Flags struct{ *flag.FlagSet }

var CommandLine = Flags{flag.CommandLine}

func New() Flags { return Flags{new(flag.FlagSet)} }

func (flags Flags) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w)
	flags.SetOutput(w)
	flags.PrintDefaults()
}

func (flags Flags) Parse(args []string) error {
	flags.Init("", flag.ContinueOnError)
	flags.Usage = func() {}
	flags.SetOutput(io.Discard)
	return flags.FlagSet.Parse(args)
}
