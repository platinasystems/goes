// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import (
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netlink/nldump"
)

const (
	Name    = "nldump"
	Apropos = "print netlink multicast messages"
	Usage   = nldump.Usage
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
	return "link | addr | route | neighor | nsid"
}

func (cmd) Main(args ...string) error { return nldump.Main(args...) }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
