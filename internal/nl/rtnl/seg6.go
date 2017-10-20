// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/sizeof"
)

const (
	SEG6_IPTUNNEL_UNSPEC uint16 = iota
	SEG6_IPTUNNEL_SRH
	N_SEG6_IPTUNNEL
)

const SEG6_IPTUNNEL_MAX = N_SEG6_IPTUNNEL - 1

const SizeofSeg6IpTunnelEncap = sizeof.Int + nl.SizeofHdr

type Seg6IpTunnelEncap struct {
	Mode int
	SrHdr
}

const SizeofSrHdr = (6 + sizeof.Byte) + sizeof.Short

type SrHdr struct {
	NextHdr      uint8
	HdrLen       uint8
	Type         uint8
	SegmentsLeft uint8
	FirstSegment uint8
	Flags        uint8
	reserved     uint16
}

const (
	SR6_FLAG1_PROTECTED uint8 = 1 << 6
	SR6_FLAG1_OAM       uint8 = 1 << 5
	SR6_FLAG1_ALERT     uint8 = 1 << 4
	SR6_FLAG1_HMAC      uint8 = 1 << 3
)

const (
	SR6_TLV_INGRESS uint8 = 1 + iota
	SR6_TLV_EGRESS
	SR6_TLV_OPAQUE
	SR6_TLV_PADDING
	SR6_TLV_HMAC
)

// SEG6_IPTUN_ENCAP_SIZE(x) ((sizeof(*x)) + (((x)->srh->hdrlen + 1) << 3))

const (
	SEG6_IPTUN_MODE_INLINE int = iota
	SEG6_IPTUN_MODE_ENCAP
)

const SEG6_IPTUN_MODE_UNSPEC int = -1

type Sr6Tlv struct {
	Type  uint8
	Len   uint8
	Value []byte
}

const SEG6_HMAC_SECRET_LEN = 64
const SEG6_HMAC_FIELD_LEN = 32

type Sr6TlvHmac struct {
	Type      uint8
	Len       uint8
	reserved  uint16
	HmacKeyId Be32
	Hmac      [SEG6_HMAC_FIELD_LEN]uint8
}

const (
	SEG6_HMAC_ALGO_SHA1   = 1
	SEG6_HMAC_ALGO_SHA256 = 2
)
