// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !cgo netgo

package sysconf

// If built with Cgo, Hz is reinitialized by sysconf(_SC_CLK_TCK).
func Hz() uint64 { return uint64(100) }
