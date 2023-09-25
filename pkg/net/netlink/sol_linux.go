// Copyright © 2015-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

const (
	SOL_IP        = 0
	SOL_SOCKET    = 1
	SOL_TCP       = 6
	SOL_UDP       = 17
	SOL_IPV6      = 41
	SOL_ICMPV6    = 58
	SOL_SCTP      = 132
	SOL_UDPLITE   = 136
	SOL_RAW       = 255
	SOL_IPX       = 256
	SOL_AX25      = 257
	SOL_ATALK     = 258
	SOL_NETROM    = 259
	SOL_ROSE      = 260
	SOL_DECNET    = 261
	SOL_X25       = 262
	SOL_PACKET    = 263
	SOL_ATM       = 264
	SOL_AAL       = 265
	SOL_IRDA      = 266
	SOL_NETBEUI   = 267
	SOL_LLC       = 268
	SOL_DCCP      = 269
	SOL_NETLINK   = 270
	SOL_TIPC      = 271
	SOL_RXRPC     = 272
	SOL_PPPOL2TP  = 273
	SOL_BLUETOOTH = 274
	SOL_PNPIPE    = 275
	SOL_RDS       = 276
	SOL_IUCV      = 277
	SOL_CAIF      = 278
	SOL_ALG       = 279
	SOL_NFC       = 280
	SOL_KCM       = 281
	SOL_TLS       = 282
	SOL_XDP       = 283
	SOL_MPTCP     = 284
	SOL_MCTP      = 285
	SOL_SMC       = 286
)

const (
	NETLINK_ADD_MEMBERSHIP = iota + 1
	NETLINK_DROP_MEMBERSHIP
	NETLINK_PKTINFO
	NETLINK_BROADCAST_ERROR
	NETLINK_NO_ENOBUFS
	NETLINK_RX_RING
	NETLINK_TX_RING
	NETLINK_LISTEN_ALL_NSID
	NETLINK_LIST_MEMBERSHIPS
	NETLINK_CAP_ACK
	NETLINK_EXT_ACK
	NETLINK_GET_STRICT_CHK
)
