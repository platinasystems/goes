// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package af

import "golang.org/x/sys/unix"

const INET = unix.AF_INET

type Inet int

var OpenInet = Open[Inet]

func (Inet) Family() int { return INET }
func (Inet) Type() int   { return unix.SOCK_DGRAM }
func (Inet) Proto() int  { return 0 }
