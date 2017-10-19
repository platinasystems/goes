// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IPPROTO_IP      uint8 = 0   // Dummy protocol for TCP
	IPPROTO_ICMP    uint8 = 1   // Internet Control Message Protocol
	IPPROTO_IGMP    uint8 = 2   // Internet Group Management Protocol
	IPPROTO_IPIP    uint8 = 4   // IPIP tunnels (older KA9Q tunnels use 94)
	IPPROTO_TCP     uint8 = 6   // Transmission Control Protocol
	IPPROTO_EGP     uint8 = 8   // Exterior Gateway Protocol
	IPPROTO_PUP     uint8 = 12  // PUP protocol
	IPPROTO_UDP     uint8 = 17  // User Datagram Protocol
	IPPROTO_IDP     uint8 = 22  // XNS IDP protocol
	IPPROTO_TP      uint8 = 29  // SO Transport Protocol Class 4
	IPPROTO_DCCP    uint8 = 33  // Datagram Congestion Control Protocol
	IPPROTO_IPV6    uint8 = 41  // IPv6-in-IPv4 tunnelling
	IPPROTO_RSVP    uint8 = 46  // RSVP Protocol
	IPPROTO_GRE     uint8 = 47  // Cisco GRE tunnels (rfc 1701,1702)
	IPPROTO_ESP     uint8 = 50  // Encapsulation Security Payload protocol
	IPPROTO_AH      uint8 = 51  // Authentication Header protocol
	IPPROTO_MTP     uint8 = 92  // Multicast Transport Protocol
	IPPROTO_BEETPH  uint8 = 94  // IP option pseudo header for BEET
	IPPROTO_ENCAP   uint8 = 98  // Encapsulation Header
	IPPROTO_PIM     uint8 = 103 // Protocol Independent Multicast
	IPPROTO_COMP    uint8 = 108 // Compression Header Protocol
	IPPROTO_SCTP    uint8 = 132 // Stream Control Transport Protocol
	IPPROTO_UDPLITE uint8 = 136 // UDP-Lite (RFC 3828)
	IPPROTO_MPLS    uint8 = 137 // MPLS in IP (RFC 4023)
	IPPROTO_RAW     uint8 = 255 // Raw IP packets

	IPPROTO_MAX = IPPROTO_RAW
)
