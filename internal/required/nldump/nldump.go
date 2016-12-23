// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import (
	"os"

	"github.com/platinasystems/go/internal/netlink"
)

const Name = "nldump"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [TYPE...]" }

func (cmd) Main(args ...string) error {
	return netlink.Dump(os.Stdout, args...)
}
