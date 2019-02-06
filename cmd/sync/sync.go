// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sync

import (
	"syscall"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "sync" }

func (Command) Usage() string { return "sync" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "flush file system buffers",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Force changed blocks to disk, update the super block.`,
	}
}

func (Command) Main(args ...string) error {
	syscall.Sync()
	return nil
}
