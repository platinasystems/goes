// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || darwin || freebsd || ios || linux || netbsd || openbsd

package af

import "syscall"

const ROUTE = syscall.AF_ROUTE

type Route int

var OpenRoute = Open[Route]

func (Route) Family() int { return ROUTE }
func (Route) Type() int   { return syscall.SOCK_RAW }
func (Route) Proto() int  { return 0 }
