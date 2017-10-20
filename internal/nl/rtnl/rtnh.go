// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"unsafe"

	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/sizeof"
)

const (
	RTNH_F_DEAD uint8 = 1 << iota // Nexthop is dead (used by multipath)

	RTNH_F_PERVASIVE  // Do recursive gateway lookup
	RTNH_F_ONLINK     // Gateway is forced on link
	RTNH_F_OFFLOAD    // offloaded route
	RTNH_F_LINKDOWN   // carrier-down on nexthop
	RTNH_F_UNRESOLVED // The entry is unresolved (ipmr)
)

const RTNH_COMPARE_MASK = RTNH_F_DEAD | RTNH_F_LINKDOWN | RTNH_F_OFFLOAD

const SizeofRtnh = sizeof.Short + sizeof.Byte + sizeof.Byte + sizeof.Int

type Rtnh struct {
	length  uint16
	Flags   uint8
	Hops    uint8
	Ifindex int
}

type RtnhAttrs struct {
	Rtnh
	nl.Attrs
}

type RtnhAttrsList []RtnhAttrs

func (v RtnhAttrs) Read(b []byte) (int, error) {
	n, err := v.Attrs.Read(b[SizeofRtnh:])
	if err != nil {
		return 0, err
	}
	n = RTNH.Align(SizeofRtnh + n)
	v.Rtnh.length = uint16(n)
	*(*Rtnh)(unsafe.Pointer(&b[0])) = v.Rtnh
	return n, nil
}

func (v RtnhAttrsList) Read(b []byte) (int, error) {
	var n, total int
	var err error
	for _, x := range v {
		n, err = x.Read(b[total:])
		if err != nil {
			break
		}
		total += n
	}
	return total, nil
}
