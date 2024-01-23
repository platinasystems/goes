// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sysctl

import (
	"github.com/platinasystems/goes/v2/pkg/align"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	Align     = align.Align(AlignTo).Roundup
	PageAlign = page.Align
	PageSize  = page.Size()
)
