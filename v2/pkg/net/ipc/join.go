// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

func join(prefix string, args []any) string {
	return fmt.Sprint(prefix, program.Base(), fmt.Sprint(args...))
}
