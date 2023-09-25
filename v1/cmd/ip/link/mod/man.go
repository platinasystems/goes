// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

const Man = `
SUBJECTS
	DEVICE	subject device

	dev DEVICE
		With "ip link add DEVICE type ... dev PHYS_DEV", the subject
		DEVICE is added to the given physical device; otherwise, "dev
		DEVICE" specifies the subject.

	link DEVICE
		physical device of the new virtual or tunnel device

	group GROUP
		With both DEVICE and "group", assigns the subject to the named
		GROUP; otherwise, perform command on the device group.

FLAGS
	Flags generally disable the named feature with a "no-" or "-" prefix
	and re-enable as named or prefaced with "+", for example,

		ip link set DEV no-arp
		ip link set DEV arp
	or
		ip link set DEV -arp
		ip link set DEV +arp

	Common flags:

	up	admin enable device
	down	disable

	no-arp	disable kernel ARP for device

	no-multicast
		disable multicast packet reception

	protodown
	no-protodown
		Set/unset device protocol error indicator so that switch
		drivers may down/up the respective switch port

OPTIONS
	name NAME
	alias NAME
		set the DEVICE name or alias

	txqueuelen NUMBER
	txqlen NUMBER
		transmit queue length

	mtu NUMBER
		Maximum Transmission Unit

	address LLADDRESS
		logical link unicast address (aka. MAC)

	broadcast LLADDRESS
	brd LLADDRESS
		logical link broadcast address

	peer LLADDRESS
		the remote's logical link address in a point-to-point connection

	master DEVICE
		controlling device

	nomaster
		disassciate with constrolling device

	addrgenmode { eui64 | none | stable_secret | random }
		IPv6 address generation mode

		eui64 - Modified EUI-64 format interface identifier
		none - disable automatic address generation
		stable_secret - generate from preset
			/proc/sys/net/ipv6/conf/{default,DEVICE}/stable_secret
		random - like stable_secret, but use random secret if no
			/proc/sys/net/ipv6/conf/{default,DEVICE}/stable_secret

	netns { NAME | PID }
		move the device to the named network namespace or to that
		associated with the identified process.

	link-netnsid ID
		peer identifier for a cross-netns interface

	vf NUMBER [ VF-ARGS ]
		see below for the VF specific arguments

	xdp [ XDP-ARGS ]
		set (or unset) a XDP ("eXpress Data Path") BPF program 

VF-ARGS
	mac LLADDRESS
		vf MAC address

	vlan VLAN-ID
		When specified, all outgoing traffic will be tagged with the
		given identifier and all incoming packets are filtered
		accordingly. A 0 VLANID disables tagging.

	qos VLAN-QOS
		VLAN QOS prority bits for all outgping traffic

	proto { [ieee]802.1q | [ieee]802.1ad }

	min-tx-rate Mbps
		Minimum transmit bandwidth

	max-tx-rate Mbps
		Maximum transmit bandwidth

	state { auto | enable | disable }
		auto - VF reflects PF's link state
		enable - VF may communicate with other VFs even if PF is down
		disable - PF drop packets sent by VF

	node-guid EUI64
	port-guid EUI64
		GUID for Infiniband VFs

	[no-]spoofchk

	[no-]query-rss
		permit/deny query of VF RSS information.

	[no-]trust
		trust VF user to set such features as multicast and promiscuous

XDP-ARGS
	off | none
		detach any currently attached XDP/BPF program

	object FILE [ section NAME ] [ [no-]verbose ]
		Attach an XDP/BPF program to the given device. FILE must be a
		BPF ELF file containing the program code, and map
		specifications. The ip '-force' option instructs the device to
		detach the current program, if any, before loading the given
		object. The default section is "prog".

	pinned FILE

SEE ALSO
	ip man link || ip link -man
	man ip || ip -man
	ip man netns || ip netns -man
	ethtool(8), iptables(8)

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
