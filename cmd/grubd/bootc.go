// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !bootrom

package grubd

import (
	mk1Bootc "github.com/platinasystems/goes/cmd/platina/mk1/bootc"
)

func bootc() []string {
	return mk1Bootc.Bootc()
}
