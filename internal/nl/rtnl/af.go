// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	AF_UNSPEC uint8 = iota
	AF_UNIX
	AF_INET
	AF_AX25
	AF_IPX
	AF_APPLETALK
	AF_NETROM
	AF_BRIDGE
	AF_ATMPVC
	AF_X25
	AF_INET6
	AF_ROSE
	AF_DECnet
	AF_NETBEUI
	AF_SECURITY
	AF_KEY
	AF_NETLINK
	AF_PACKET
	AF_ASH
	AF_ECONET
	AF_ATMSVC
	AF_RDS
	AF_SNA
	AF_IRDA
	AF_PPPOX
	AF_WANPIPE
	AF_LLC
	AF_IB
	AF_MPLS
	AF_CAN
	AF_TIPC
	AF_BLUETOOTH
	AF_IUCV
	AF_RXRPC
	AF_ISDN
	AF_PHONET
	AF_IEEE802154
	AF_CAIF
	AF_ALG
	AF_NFC
	AF_VSOCK
)

var AfByName = map[string]uint8{
	"unspec":     AF_UNSPEC,
	"unix":       AF_UNIX,
	"inet":       AF_INET,
	"ax25":       AF_AX25,
	"ipx":        AF_IPX,
	"appletalk":  AF_APPLETALK,
	"netrom":     AF_NETROM,
	"bridge":     AF_BRIDGE,
	"atmpvc":     AF_ATMPVC,
	"x25":        AF_X25,
	"inet6":      AF_INET6,
	"rose":       AF_ROSE,
	"decnet":     AF_DECnet,
	"netbeui":    AF_NETBEUI,
	"security":   AF_SECURITY,
	"key":        AF_KEY,
	"netlink":    AF_NETLINK,
	"packet":     AF_PACKET,
	"ash":        AF_ASH,
	"econet":     AF_ECONET,
	"atmsvc":     AF_ATMSVC,
	"rds":        AF_RDS,
	"sna":        AF_SNA,
	"irda":       AF_IRDA,
	"pppox":      AF_PPPOX,
	"wanpipe":    AF_WANPIPE,
	"llc":        AF_LLC,
	"ib":         AF_IB,
	"mpls":       AF_MPLS,
	"can":        AF_CAN,
	"tipc":       AF_TIPC,
	"bluetooth":  AF_BLUETOOTH,
	"iucv":       AF_IUCV,
	"rxrpc":      AF_RXRPC,
	"isdn":       AF_ISDN,
	"phonet":     AF_PHONET,
	"ieee802154": AF_IEEE802154,
	"caif":       AF_CAIF,
	"alg":        AF_ALG,
	"nfc":        AF_NFC,
	"vsock":      AF_VSOCK,
}

func AfName(af uint8) string {
	s, found := map[uint8]string{
		AF_UNSPEC:     "unspec",
		AF_UNIX:       "unix",
		AF_INET:       "inet",
		AF_AX25:       "ax25",
		AF_IPX:        "ipx",
		AF_APPLETALK:  "appletalk",
		AF_NETROM:     "netrom",
		AF_BRIDGE:     "bridge",
		AF_ATMPVC:     "atmpvc",
		AF_X25:        "x25",
		AF_INET6:      "inet6",
		AF_ROSE:       "rose",
		AF_DECnet:     "decnet",
		AF_NETBEUI:    "netbeui",
		AF_SECURITY:   "security",
		AF_KEY:        "key",
		AF_NETLINK:    "netlink",
		AF_PACKET:     "packet",
		AF_ASH:        "ash",
		AF_ECONET:     "econet",
		AF_ATMSVC:     "atmsvc",
		AF_RDS:        "rds",
		AF_SNA:        "sna",
		AF_IRDA:       "irda",
		AF_PPPOX:      "pppox",
		AF_WANPIPE:    "wanpipe",
		AF_LLC:        "llc",
		AF_IB:         "ib",
		AF_MPLS:       "mpls",
		AF_CAN:        "can",
		AF_TIPC:       "tipc",
		AF_BLUETOOTH:  "bluetooth",
		AF_IUCV:       "iucv",
		AF_RXRPC:      "rxrpc",
		AF_ISDN:       "isdn",
		AF_PHONET:     "phonet",
		AF_IEEE802154: "ieee802154",
		AF_CAIF:       "caif",
		AF_ALG:        "alg",
		AF_NFC:        "nfc",
		AF_VSOCK:      "vsock",
	}[af]
	if !found {
		s = "unknown"
	}
	return s
}

var AfBits = map[uint8]uint8{
	AF_INET:  32,
	AF_INET6: 128,
	AF_MPLS:  20,
}
