// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nlcounters

import "github.com/platinasystems/go/internal/netlink/nlcounters"

const Name = "nlcounters"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return nlcounters.Usage }

func (cmd) Main(args ...string) error {
	return nlcounters.Main(args...)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "periodic print of netlink interface counters",
	}
}

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
