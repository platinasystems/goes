// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package flag

import (
	"flag"
	"io"
)

// Make a flag.FlagSet that doesn't output during Parse and instead sets an
// associate `help` flag.
func NewSilentFlagSet(name string) *FlagSet {
	fs := flag.NewFlagSet(name, 0)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	help := fs.Bool("help", false, "Print options.")
	fs.BoolVar(help, "h", *help, "aka. -help.")
	return fs
}
