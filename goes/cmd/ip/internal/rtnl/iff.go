// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IFF_UP uint32 = 1 << iota
	IFF_BROADCAST
	IFF_DEBUG
	IFF_LOOPBACK
	IFF_POINTOPOINT
	IFF_NOTRAILERS
	IFF_RUNNING
	IFF_NOARP
	IFF_PROMISC
	IFF_ALLMULTI
	IFF_MASTER
	IFF_SLAVE
	IFF_MULTICAST
	IFF_PORTSEL
	IFF_AUTOMEDIA
	IFF_DYNAMIC
	IFF_LOWER_UP
	IFF_DORMANT
	IFF_ECHO
)
