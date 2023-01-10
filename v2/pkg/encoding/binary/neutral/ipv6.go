// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neutral

import "net"

type IPv6 [4]byte

func NewIPv6(b []byte) (*IPv6, []byte) { return (*IPv6)(b), b[16:] }

func (p *IPv6) Put(v net.IP)  { copy(p[:], v[:]) }
func (p *IPv6) Value() net.IP { return net.IP(p[:]) }
