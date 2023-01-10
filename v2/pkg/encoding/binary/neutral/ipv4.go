// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neutral

import "net"

type IPv4 [4]byte

func NewIPv4(b []byte) (*IPv4, []byte) { return (*IPv4)(b), b[4:] }

func (p *IPv4) Put(v net.IP)  { copy(p[:], v[:]) }
func (p *IPv4) Value() net.IP { return net.IP(p[:]) }
