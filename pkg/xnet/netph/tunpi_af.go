// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin

package netph

import "github.com/platinasystems/goes/v2/pkg/xnet"

const (
	TUN_P_IP  = xnet.AF_INET
	TUN_P_IP6 = xnet.AF_INET6
)
