// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

func main() {
	goes.Root = map[string]any{
		"build":    program.Build,
		"cat":      cat.Func,
		"command":  command.Func,
		"echo":     echo.Func,
		"hostname": host.Name,
		"main":     program.Main,
	}
	goes.Main()
}
