// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sysctl

import "github.com/platinasystems/goes/v2/pkg/syscall/align"

var (
	Align     = align.Align(AlignTo).Roundup
	PageAlign = align.Page.Roundup
	PageSize  = align.Page.Size
)
