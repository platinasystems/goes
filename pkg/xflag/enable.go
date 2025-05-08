// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"flag"
	"strconv"
)

// [EnableIn] [flag.CommandLine]
func Enable(name, usage string, f func() error) {
	EnableIn(flag.CommandLine, name, usage, f)
}

// EnableIn defines a boolean flag that if parsed true, calls “f”.
func EnableIn(fs *flag.FlagSet, name, usage string, f func() error) {
	fs.Var(enabling{f}, name, usage)
}

type enabling struct{ f func() error }

func (enabling) IsBoolFlag() bool { return true }
func (enabling) String() string   { return "false" }

func (r enabling) Set(s string) error {
	if len(s) > 0 {
		if t, err := strconv.ParseBool(s); !t || err != nil {
			return err
		}
	}
	return r.f()
}
