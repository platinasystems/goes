// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

const Man = `
DESCRIPTION
   ip address show - look at protocol addresses
	dev IFNAME (default)
	      name of device.

	scope SCOPE_VAL
	      only list addresses with this scope.

	to PREFIX
	      only list addresses matching this prefix.

	label PATTERN
	      only list addresses with labels matching the PATTERN.  PATTERN
	      is a usual shell style pattern.

	master DEVICE
	      only list interfaces enslaved to this master device.

	vrf NAME
	      only list interfaces enslaved to this vrf.

	type TYPE
	      only list interfaces of the given type.

	      Note that the type name is not checked against the list of sup‐
	      ported types - instead it is sent as-is to the kernel. Later it
	      is used to filter the returned interface list by comparing it
	      with the relevant attribute in case the kernel didn't filter
	      already. Therefore any string is accepted, but may lead to empty
	      output.

	up     only list running interfaces.

	dynamic and permanent
	      (IPv6 only) only list addresses installed due to stateless
	      address configuration or only list permanent (not dynamic)
	      addresses.

	tentative
	      (IPv6 only) only list addresses which have not yet passed dupli‐
	      cate address detection.

	-tentative
	      (IPv6 only) only list addresses which are not in the process of
	      duplicate address detection currently.

	deprecated
	      (IPv6 only) only list deprecated addresses.

	-deprecated
	      (IPv6 only) only list addresses not being deprecated.

	dadfailed
	      (IPv6 only) only list addresses which have failed duplicate
	      address detection.

	-dadfailed
	      (IPv6 only) only list addresses which have not failed duplicate
	      address detection.

	temporary
	      (IPv6 only) only list temporary addresses.

	primary and secondary
	      only list primary (or secondary) addresses.

EXAMPLES
	ip address show
	   Shows IPv4 and IPv6 addresses assigned to all network interfaces.
	   The 'show' subcommand can be omitted.

	ip address show up
	   Same as above except that only addresses assigned to active network
	   interfaces are shown.

	ip address show dev eth0
	   Shows IPv4 and IPv6 addresses assigned to network interface eth0.

SEE ALSO
	ip man address || ip address -man
	man ip || ip -man

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
