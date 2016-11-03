/*
Copyright 2015-2016 Platina Systems, Inc. All rights reserved.  Use of this
source code is governed by a BSD-style license described in the LICENSE file.
*/

/*
Package machined provides a goes daemon to detect, configure, and publish
the machine info (eg. cards, fans, and netdevs) through redis.

	hkeys platina
		returns the machine's attributes

	hgetall platina
		returns a list of all machine attributes as ATTR, VALUE pairs

	hget platina ATTR
		returns the specified machine attribute

	hget platina DEV.ATTR
		returns the specified device attribute

	hset platina DEV CIDR[;ATTR[=VALUE]]...
		add an address to the device

	hdel platina DEV.CIDR
		removes the previously set address from device
		NOTE: ip addresses must be encapsulated with brackets

	hset platina DEV.ATTR VALUE
		sets the specified device attribute

e.g.
	hset platina fan.front 50
	hset platina eth0 10.10.10.10/32
	hset platina eth0 10.0.0.1/24;broadcast=10.0.0.255
	hset platina eth0.mtu 9140
	hset platina eth0.admin down
	hdel platina eth0.[10.10.10.10/32]

Configurable Device Attributes:

	admin		bool
	debug		bool
	no-trailers	bool
	no-arp		bool
	promiscuous	bool
	all-multicast	bool
	multicast-encap	bool
	port-select	bool
	automedia	bool
	dynamic		bool
	mtu		int

Configurable Device Address Attributes:

	broadcast	ip or ipv6 address
	local		TBD
	label		TBD
	anycast		TBD

Machines may customize info providers within their main() by,

  * adding network devices

	machined.NetLink.Prefixes("lo.", "eth0.")

  * adding info providers

	machined.InfoProviders = append(machined.InfoProviders,
		CustomProviders...)

  * overwriting an existing provider

	machined.Hostname = CustomHostname
*/
package machined
