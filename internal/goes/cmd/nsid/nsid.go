// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nsid

import (
	"io/ioutil"
	"strings"

	"github.com/platinasystems/go/internal/netlink/nsid"
)

const Name = "nsid"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return nsid.Usage }

func (cmd) Main(args ...string) error {
	return nsid.Main(args...)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "net namespace identifier config",
	}
}

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

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	nsid - net namespace identifier config

SYNOPSIS
	nsid [list [NAME]...]
	nsid set NAME ID
	nsid unet NAME ID

DESCRIPTION
	[list [NAME]...]
		show the identifier of each network namespace with "-1"
		indicating an unidentifeid namespace.

	set	set the namespace identifier

	unset	unset the namespace identifier`,
	}
}
