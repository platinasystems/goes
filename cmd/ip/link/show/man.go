// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

const Man = `
DESCRIPTION
   ip link show - display device attributes
	dev NAME (default)
	       NAME specifies the network device to show.  If this argument is
	       omitted all devices in the default group are listed.

	group GROUP
	       GROUP specifies what group of devices to show.

	up     only display running interfaces.

	master DEVICE
	       DEVICE specifies the master device which enslaves devices to
	       show.

	vrf NAME
	       NAME speficies the VRF which enslaves devices to show.

	type TYPE
	       TYPE specifies the type of devices to show.

	       Note that the type name is not checked against the list of sup‐
	       ported types - instead it is sent as-is to the kernel. Later it
	       is used to filter the returned interface list by comparing it
	       with the relevant attribute in case the kernel didn't filter
	       already. Therefore any string is accepted, but may lead to empty
	       output.

   ip link help - display help
	TYPE specifies which help of link type to dislpay.

   GROUP
	may be a number or a string from the file /etc/iproute2/group which can
	be manually filled.

EXAMPLES
	ip link show
	    Shows the state of all network interfaces on the system.

	ip link show type bridge
	    Shows the bridge devices.

	ip link show type vlan
	    Shows the vlan devices.

	ip link show master br0
	    Shows devices enslaved by br0

SEE ALSO
	ip man link || ip link -man
	man ip || ip -man
	ip man netns || ip netns -man
	ethtool(8), iptables(8)

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
