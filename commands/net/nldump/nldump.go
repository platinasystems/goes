// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package nldump

import (
	"os"

	"github.com/platinasystems/go/netlink"
)

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return "nldump" }
func (cmd) Usage() string  { return "nldump [TYPE...]" }

func (cmd) Main(args ...string) error {
	return netlink.Dump(os.Stdout, args...)
}
