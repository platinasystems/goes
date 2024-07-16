// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "golang.org/x/sys/unix"

const AF_SYSTEM = unix.AF_SYSTEM

type System int

var OpenSystem = Open[System]

func (System) Family() int { return AF_SYSTEM }
func (System) Type() int   { return unix.SOCK_DGRAM }
func (System) Proto() int  { return 0 }
