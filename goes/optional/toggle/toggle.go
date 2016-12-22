// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package toggle

import (
	"fmt"
	"github.com/platinasystems/go/goes/optional/i2c"
)

const Name = "toggle"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " SECONDS" }

func (cmd) Main(args ...string) error {
	x, _ := i2c.ReadByte(0, 0x74, 2)
	if (x != 0xdf) && (x != 0xff) {
		return fmt.Errorf("x86 not ready, no toggle")
	}
	x ^= 0x20
	i2c.WriteByte(0, 0x74, 2, x)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "toggle console port between x86 and BMC",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	toggle - toggle console port between x86 and BMC

SYNOPSIS
	toggle

DESCRIPTION
	The toggle command toggles the console port between x86 and BMC.`,
	}
}
