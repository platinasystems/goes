// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const SOL_NETLINK = 270

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
)
