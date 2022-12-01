// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

var Service = selection.Map{
	"cat":     cat.Func,
	"command": command.Func,
	"echo":    echo.Func,
}
