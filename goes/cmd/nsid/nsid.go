// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nsid

import (
	"io/ioutil"
	"strings"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netlink/nsid"
)

const (
	Name    = "nsid"
	Apropos = "net namespace identifier config"
	Usage   = nsid.Usage
	Man     = `
DESCRIPTION
	[list [NAME]...]
		show the identifier of each network namespace with "-1"
		indicating an unidentifeid namespace.

	set	set the namespace identifier

	unset	unset the namespace identifier`
)

type Interface interface {
	Apropos() lang.Alt
	Complete(...string) []string
	Help(...string) string
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Complete(args ...string) (c []string) {
	var cmds = []string{
		"list",
		"set",
		"unset",
	}
	if len(args) > 0 && strings.HasSuffix(args[0], "nsid") {
		args = args[1:]
	}
	switch len(args) {
	case 0:
		c = cmds
	case 1:
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, args[0]) {
				c = append(c, cmd)
			}
		}
	default:
		lastarg := args[len(args)-1]
		dir, err := ioutil.ReadDir(nsid.VarRunNetns)
		if err != nil {
			break
		}
		for _, info := range dir {
			name := info.Name()
			if strings.HasPrefix(name, lastarg) {
				c = append(c, name)
			}
		}
	}
	return
}

func (cmd) Help(args ...string) string {
	help := "no help"
	switch {
	case len(args) == 0:
		help = "list | set | unset"
	case args[0] == "list":
		return "NAME"
	case args[0] == "set", args[0] == "unset":
		switch len(args) {
		case 1:
			help = "NAME"
		case 2:
			help = "ID"
		}
	}
	return help
}

func (cmd) Main(args ...string) error { return nsid.Main(args...) }

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
