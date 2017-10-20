// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

/*
Reference: RFC 5462, RFC 3032

	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                Label                  | TC  |S|       TTL     |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	Label:  Label Value, 20 bits
	TC:     Traffic Class field, 3 bits
	S:      Bottom of Stack, 1 bit
	TTL:    Time to Live, 8 bits
*/

const (
	MPLS_LS_LABEL_MASK uint32 = 0xFFFFF000
	MPLS_LS_TC_MASK    uint32 = 0x00000E00
	MPLS_LS_S_MASK     uint32 = 0x00000100
	MPLS_LS_TTL_MASK   uint32 = 0x000000FF

	MPLS_LS_LABEL_SHIFT = 12
	MPLS_LS_TC_SHIFT    = 9
	MPLS_LS_S_SHIFT     = 8
	MPLS_LS_TTL_SHIFT   = 0
)

// Reserved labels
const (
	MPLS_LABEL_IPV4NULL  uint32 = 0  // RFC3032
	MPLS_LABEL_RTALERT   uint32 = 1  // RFC3032
	MPLS_LABEL_IPV6NULL  uint32 = 2  // RFC3032
	MPLS_LABEL_IMPLNULL  uint32 = 3  // RFC3032
	MPLS_LABEL_ENTROPY   uint32 = 7  // RFC6790
	MPLS_LABEL_GAL       uint32 = 13 // RFC5586
	MPLS_LABEL_OAMALERT  uint32 = 14 // RFC3429
	MPLS_LABEL_EXTENSION uint32 = 15 // RFC7274

	MPLS_LABEL_FIRST_UNRESERVED uint32 = 16 // RFC3032
)

var MplsReservedLabelByName = map[string]uint32{
	"ipv4null":  MPLS_LABEL_IPV4NULL,
	"rtalert":   MPLS_LABEL_RTALERT,
	"ipv6null":  MPLS_LABEL_IPV6NULL,
	"implnull":  MPLS_LABEL_IMPLNULL,
	"entropy":   MPLS_LABEL_ENTROPY,
	"gal":       MPLS_LABEL_GAL,
	"oamalert":  MPLS_LABEL_OAMALERT,
	"extension": MPLS_LABEL_EXTENSION,

	"first-unreserved": MPLS_LABEL_FIRST_UNRESERVED,
}

const (
	MPLS_IPTUNNEL_UNSPEC uint16 = iota
	MPLS_IPTUNNEL_DST
	MPLS_IPTUNNEL_TTL
	N_MPLS_IPTUNNEL
)

const MPLS_IPTUNNEL_MAX = N_MPLS_IPTUNNEL - 1
