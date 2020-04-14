// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package version

import (
	"fmt"

	"github.com/platinasystems/goes/internal/buildinfo"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "version" }
func (Command) Usage() string  { return "[show ]version" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print goes-MACHINE version",
	}
}

func (Command) Main(...string) error {
	fmt.Println(buildinfo.New().Version())
	return nil
}
