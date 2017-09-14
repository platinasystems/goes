// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	RTN_UNSPEC uint8 = iota
	RTN_UNICAST
	RTN_LOCAL
	RTN_BROADCAST
	RTN_ANYCAST
	RTN_MULTICAST
	RTN_BLACKHOLE
	RTN_UNREACHABLE
	RTN_PROHIBIT
	RTN_THROW
	RTN_NAT
	RTN_XRESOLVE
)

var RtnByName = map[string]uint8{
	"unspec":      RTN_UNSPEC,
	"unicast":     RTN_UNICAST,
	"local":       RTN_LOCAL,
	"broadcast":   RTN_BROADCAST,
	"brd":         RTN_BROADCAST,
	"anycast":     RTN_ANYCAST,
	"multicast":   RTN_MULTICAST,
	"blackhole":   RTN_BLACKHOLE,
	"unreachable": RTN_UNREACHABLE,
	"prohibit":    RTN_PROHIBIT,
	"throw":       RTN_THROW,
	"nat":         RTN_NAT,
	"xresolve":    RTN_XRESOLVE,
}
