// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmdline

import (
	"fmt"

	"github.com/platinasystems/go/internal/cmdline"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "cmdline"
	Apropos = "parse and print /proc/cmdline variables"
	Usage   = "cmdline [NAME]..."
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	keys, m, err := cmdline.New()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		keys = args
	}
	for _, k := range keys {
		if _, found := m[k]; !found {
			return fmt.Errorf("%s: not found", k)
		}
	}
	if len(keys) == 1 {
		fmt.Println(m[keys[0]])
	} else {
		for _, k := range keys {
			fmt.Print(k, ": ", m[k], "\n")
		}
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
