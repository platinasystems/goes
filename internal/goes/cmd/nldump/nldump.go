// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import "github.com/platinasystems/go/internal/netlink/nldump"

const Name = "nldump"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return nldump.Usage }

func (cmd) Main(args ...string) error {
	return nldump.Main(args...)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print netlink multicast messages",
	}
}

func (cmd) Help(args ...string) string {
	return "link | addr | route | neighor | nsid"
}
