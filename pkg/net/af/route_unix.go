// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package af

import "golang.org/x/sys/unix"

const ROUTE = unix.AF_ROUTE

type Route int

var OpenRoute = Open[Route]

func (Route) Family() int { return ROUTE }
func (Route) Type() int   { return unix.SOCK_RAW }
func (Route) Proto() int  { return 0 }
