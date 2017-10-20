// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"

	"github.com/platinasystems/go/internal/nl"
)

const RTA = nl.NLATTR
const RTNH nl.Align = syscall.RTNH_ALIGNTO
