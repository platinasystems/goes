// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nlcounters

import (
	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/netlink/nlcounters"
)

const (
	Name    = "nlcounters"
	Apropos = "periodic print of netlink interface counters"
	Usage   = nlcounters.Usage
)

type Interface interface {
	Apropos() lang.Alt
	Help(...string) string
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Help(args ...string) string {
	help := "no help"
	switch {
	case len(args) == 0:
		help = "[-i SECONDS]"
	case args[0] == "-i":
		return "SECONDS"
	}
	return help
}

func (cmd) Main(args ...string) error { return nlcounters.Main(args...) }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
