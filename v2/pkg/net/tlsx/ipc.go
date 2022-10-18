// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/net/ipc"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	base   = program.Base.String()
	IPC    = ipc.New(base)
	RegIPC = ipc.New(fmt.Sprint(base, ".registry"))
)
