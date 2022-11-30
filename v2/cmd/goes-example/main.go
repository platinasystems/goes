// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/show"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

func main() {
	selection.Map{
		"cat":     cat.Func,
		"command": command.Func,
		"echo":    echo.Func,
		"show": selection.Map{
			"build-id":   show.New(program.BuildId),
			"build-info": show.New(program.BuildInfo),
			"version":    show.New(program.MainVersion),
		}.Select,
	}.Main()
}
