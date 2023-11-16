// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin

package af

import "syscall"

const SYSTEM = syscall.AF_SYSTEM

type System int

var OpenSystem = Open[System]

func (System) Family() int { return SYSTEM }
func (System) Type() int   { return syscall.SOCK_DGRAM }
func (System) Proto() int  { return 0 }
