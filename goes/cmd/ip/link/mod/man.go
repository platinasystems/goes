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

	type TYPE [ TYPE-ARGS ]
		see below for TYPES and the TYPE specific arguments

	vf NUMBER [ VF-ARGS ]
		see below for the VF specific arguments

	xdp [ XDP-ARGS ]
		set (or unset) a XDP ("eXpress Data Path") BPF program 
		
TYPE
	Must be one of these types of virtual or associate devices:

		bridge - Ethernet Bridge device
		bond - Bonding device
		can - Controller Area Network interface
		dummy - Dummy network interface
		hsr - High-availability Seamless Redundancy device
		ifb - Intermediate Functional Block device
		ipoib - IP over Infiniband device
		macvlan - Virtual interface base on link layer address (MAC)
		macvtap - Virtual based on link layer address (MAC) and TAP.
		vcan - Virtual Controller Area Network interface
		veth - Virtual ethernet interface
		vlan - 802.1q tagged virtual LAN interface
		vxlan - Virtual eXtended LAN
		ip6tnl - Virtual tunnel interface IPv4|IPv6 over IPv6
		ipip - Virtual tunnel interface IPv4 over IPv4
		sit - Virtual tunnel interface IPv6 over IPv4
		gre - Virtual tunnel interface GRE over IPv4
		gretap - Virtual L2 tunnel interface GRE over IPv4
		ip6gre - Virtual tunnel interface GRE over IPv6
		ip6gretap - Virtual L2 tunnel interface GRE over IPv6
		vti - Virtual tunnel interface
		nlmon - Netlink monitoring device
		ipvlan - L3 (IPv6/IPv4) based VLANs
		lowpan - 6LoWPAN (IPv6) over IEEE 802.15.4 / Bluetooth
		geneve - GEneric NEtwork Virtualization Encapsulation
		macsec - IEEE 802.1AE MAC Security (MAC‐sec)
		vrf - Interface for L3 VRF domains

	Arguments common to all types:

	numtxqueues COUNT
		number of transmit queues for new device.

	numrxqueues COUNT
		number of receive queues for new device.

	index IDX
		index of the new virtual device. (fails if busy)

TYPE VLAN
	ip link add link DEVICE name NAME type vlan id ID
		[ protocol { 802.1q or 802.1ad } ]
		[ [no-]reorder-hdr ]
		[ [no-]gvrp ]
		[ [no-]mvrp ]
		[ [no-]loose-binding ]
		[ ingress-qos-map MAP ]
		[ egress-qos-map MAP ]

       no-reorder-hdr - disable reordered ethernet headers
       reorder-hdr - re-enable reordered ethernet headers

		With reorder-hdr, the VLAN header isn't inserted immediately
		but only before passing to the physical device (if this device
		does not support VLAN offloading), the similar on the RX
		direction - by default the packet will be untagged before being
		received by VLAN device. Reordering allows to accel‐ erate
		tagging on egress and to hide VLAN header on ingress so the
		packet looks like regular Ethernet packet, at the same time it
		might be confusing for packet capture as the VLAN header does
		not exist within the packet.

		VLAN offloading can be checked by ethtool(8):

			ethtool -k <phy_dev> | grep tx-vlan-offload

		Where <phy_dev> is the physical device to which VLAN
		device is bound.

	[no-]gvrp
		GARP VLAN Registration Protocol

	[no-]mvrp
		Multiple VLAN Registration Protocol

	[no-]loose-binding
		bond VLAN to the physical device state

	ingress-qos-map QOS-MAP
		defines a mapping of VLAN header prio field to the Linux
		internal packet priority on incoming frames. The format is
		FROM:TO with multiple mappings separated by spaces.

	egress-qos-map QOS-MAP
		defines a mapping of Linux internal packet priority to VLAN
		header prio field but for outgoing frames. The format is the
		same as for ingress-qos-map.

		Linux packet priority can be set by iptables(8):

			iptables -t mangle -A POSTROUTING [...]
				-j CLASSIFY --set-class 0:4

		and this "4" priority can be used in the egress qos
		mapping to set VLAN prio "5":

			ip link set veth0.10 type vlan egress 4:5

TYPE VXLAN
	ip link add [link] DEVICE type vxlan id VNI
	       [ dev PHYS_DEV  ]
	       [ { group | remote } IPADDR ]
	       [ local { IPADDR | any } ]
	       [ ttl TTL ]
	       [ tos TOS ]
	       [ flowlabel FLOWLABEL ]
	       [ dstport PORT ]
	       [ srcport MIN MAX ]
	       [ [no-]learning ]
	       [ [no-]proxy ]
	       [ [no-]rsc ]
	       [ [no-]l2miss ]
	       [ [no-]l3miss ]
	       [ [no-]udpcsum ]
	       [ [no-]udp6zerocsumtx ]
	       [ [no-]udp6zerocsumrx ]
	       [ ageing SECONDS ]
	       [ maxaddress NUMBER ]
	       [ [no-]external ]
	       [ [no-]gbp ]
	       [ [no-]gpe ]

	id VNI	VXLAN Network Identifer (or VXLAN Segment Identifier)

	dev PHYS_DEV
	       physical device of tunnel endpoint

	group IPADDR
		IP multicast address ("remote" exclusive)

	remote IPADDR
		IP unicast destination address when the link layer address is
		unknown in the VXLAN device forwarding database. ("group"
		exclusive)

	local IPADDR
		IP source address

	ttl TTL
		Time-to-live of transimitted packets

	tos TOS
		Type-of-service of transimitted packets

	flowlabel FLOWLABEL
		of transimitted packets

	dstport PORT
		UDP destination port of remote VXLAN tunnel endpoint

	srcport MIN:MAX
		range of UDP source port numbers of transimetted packets

	[no-]learning
		specifies whether unknown source link layer addresses and IP
		addresses are entered into the VXLAN device forwarding database
		(FDB)

	[no-]rsc
		route short circuit

	[no-]proxy
		proxy ARP

	[no-]l2miss
		netlink LLADDRESS miss notifications

	[no-]l3miss
		netlink IP ADDR miss notifications

	[no-]udpcsum
		UDP checksum of IPv4 transimitted packets

	[no-]udp6zerocsumtx
		UDP checksum calculation of IPv6 transmitted packets

	[no-]udp6zerocsumrx
		accept IPv6 packets w/o UDP checksum

	ageing SECONDS
		FDB lifetime of entries learnt by the kernel.

	maxaddress NUMBER
		maximum number of FDB entries

	[no-]external
		external control plane (e.g. ip route encap);
		otherwise, kernel internal FDB

	[no-]gbp
		Group Policy extension (VXLAN-GBP).

		Allows to transport group policy context across VXLAN network
		peers.  If enabled, includes the mark of a packet in the VXLAN
		header for outgoing packets and fills the packet mark based on
		the information found in the VXLAN header for incomming
		packets.

		Format of upper 16 bits of packet mark (flags);

			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			|-|-|-|-|-|-|-|-|-|D|-|-|A|-|-|-|
			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

		D := Don't Learn bit.
		When set, this bit indicates that the egress VTEP MUST NOT
		learn the source address of the encapsulated frame.

		A := Indicates that the group policy has already
		been applied to this packet. Policies MUST NOT be
		applied by devices when the A bit is set.

		Format of lower 16 bits of packet mark (policy ID):

			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			|        Group Policy ID        |
			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

		Example:
			iptables -A OUTPUT [...] -j MARK --set-mark 0x800FF

	[no-]gpe
		Generic Protocol extension (VXLAN- GPE)
		Only supported with the external control plane.

	GRE, IPIP, SIT Type Support
	       For a link of types GRE/IPIP/SIT the following additional argu‐
	       ments are supported:

	       ip link add DEVICE type { gre | ipip | sit }
	           remote ADDR local ADDR
		       [ encap { fou | gue | none } ]
		       [ encap-sport { PORT | auto } ]
		       [ encap-dport PORT ]
		       [ [no-]encap-csum ]
		       [ [no-]encap-remcsum ]

	               remote ADDR - specifies the remote address of the tun‐
	               nel.

	               local ADDR - specifies the fixed local address for tun‐
	               neled packets.  It must be an address on another inter‐
	               face on this host.

	               encap { fou | gue | none } - specifies type of secondary
	               UDP encapsulation. "fou" indicates Foo-Over-UDP, "gue"
	               indicates Generic UDP Encapsulation.

	               encap-sport { PORT | auto } - specifies the source port
	               in UDP encapsulation.  PORT indicates the port by num‐
	               ber, "auto" indicates that the port number should be
	               chosen automatically (the kernel picks a flow based on
	               the flow hash of the encapsulated packet).

	               [no-]encap-csum - specifies if UDP checksums are enabled
	               in the secondary encapsulation.

	               [no]-encap-remcsum - specifies if Remote Checksum
		       Offload is enabled. This is only applicable for Generic
		       UDP Encapsulation.

TYPE IP6GRE/IP6GRETAP
	ip link add DEVICE type { ip6gre | ip6gretap } remote ADDR local ADDR
		[ [no-]{i|o}seq ]
		[ {i|o}key KEY | no-{i|o}key ]
		[ [i|o]csum ]
		[ hoplimit TTL ]
		[ encaplimit ELIM ]
		[ tclass TCLASS ]
		[ flowlabel FLOWLABEL ]
		[ dscp inherit ]
		[ dev PHYS_DEV ]

	remote ADDR
		IPv6 address or tunnel's remote end-point

	local ADDR
		IPv6 address or tunnel's local end-point

	[no-]{i|o}seq
		enable/disable sequencing of incomming and outgoing packets.

	key KEY
	ikey KEY
	okey KEY
	no-ikey
	no-okey
		The GRE KEY may be a number or an IPv4 address-like dotted
		quad.  The "key" parameter uses the same key for both input and
		output.  The "ikey" and "okey" parameters specify different
		keys for input and output. "no-ikey" and "no-okey" remove the
		respective keys.

	[no-]csum
	[no-]icsum
	[no-]ocsum
		Validate/generate checksums for tunneled packets.  The "ocsum"
		or "csum" flags calculate checksums for outgoing packets.  The
		"icsum" or "csum" flag validates the checksum of incoming
		packets have the correct checksum.

	hoplimit TTL
		Hop Limit of outgoing packets

	encaplimit ELIM
		Fixed encapsulation limit (default, 4)

	flowlabel FLOWLABEL
		fixed flowlabel

	tclass TCLASS
		traffic class of tunneled packets, which can be specified as
		either a two-digit hex value (e.g. c0) or a predefined string
		(e.g. internet).  The value inherit causes the field to be
		copied from the original IP header. The values inherit/STRING
		or inherit/00..ff will set the field to STRING or 00..ff when
		tunneling non-IP packets. The default value is 00.

TYPE IPOIB
	ip link add DEVICE name NAME type ipoib [ pkey PKEY ] [ mode MODE ]

	pkey PKEY
		IB P-Key

	mode { datagram | connected }

TYPE GENEVE
	ip link add DEVICE type geneve id VNI remote IPADDR [ ttl TTL ]
	       [ tos TOS ] [ flowlabel FLOWLABEL ]

	id VNI
		Virtual Network Identifer

	remote IPADDR 
		IP unicast destination of outgoing packets

	ttl TTL
		Time-to-live of outgoing packets

	tos TOS
		Type-of-service of outgoing packets

	flowlabel FLOWLABEL
		Flow label of outgoing packets.

TYPE MACVLAN and MACVTAP
	ip link add link DEVICE name NAME type { macvlan | macvtap }
		mode { private | vepa | bridge | passthru  [ nopromisc ] }

	type { macvlan | macvtap } 
		macvlan just creates a virtual interface, while macvtap also
		make a character device /dev/tapX to be used just like tuntap.

	modes:
		private
			Do not allow communication between macvlan instances on
			the same physical interface, even if the external
			switch supports hairpin mode.

		vepa	Virtual Ethernet Port Aggregator mode. Data from one
			macvlan instance to the other on the same phys‐ ical
			interface is transmitted over the physical inter‐ face.
			Either the attached switch needs to support hair‐ pin
			mode, or there must be a TCP/IP router forwarding the
			packets in order to allow communication. This is the
			default mode.

		bridge	In bridge mode, all endpoints are directly connected to
			each other, communication is not redirected through the
			physical interface's peer.

		passthru [ nopromisc ]
			This mode gives more power to a single endpoint,
			usually in macvtap mode. It is not allowed for more
			than one endpoint on the same physical interface. All
			traffic will be forwarded to this end‐ point, allowing
			virtio guests to change MAC address or set promiscuous
			mode in order to bridge the interface or create vlan
			interfaces on top of it. By default, this mode forces
			the underlying interface into promiscuous mode. Passing
			the nopromisc flag prevents this, so the promisc flag
			may be controlled using standard tools.

TYPE HSR
	ip link add link DEVICE name NAME type hsr slave1 SLAVE1-IF
		slave2 SLAVE2-IF [ supervision ADDR-BYTE ]
		[ version { 0 | 1 } ]

	slave1 SLAVE1-IF
	slave2 SLAVE2-IF
		physical device for the first and second ring ports.

	supervision ADDR-BYTE
		The last byte of the multicast address for HSR supervision
		frames.  Default option is "0", possible values 0-255.

	version { 0 | 1 }
		Default, "0" corresponds to the 2010 version of the HSR
		standard. Option "1" activates the 2012 version.

TYPE MACSEC
	ip link add link DEVICE name NAME type macsec
		[ port PORT | sci SCI ]
		[ cipher CIPHER_SUITE ]
		[ icvlen { 8..16 } ]
		[ [no-]encrypt ] [ [no-]send_sci ] [ [no-]end_station ]
		[ [no-]scb ] [ [no-]protect ] [ [no-]replay ] 
		[ window { 0..2^32-1 } ]
		[ validate { strict | check | disabled } ]
		[ encodingsa { 0..3 } ]

	port PORT

	sci SCI

	cipher CIPHER_SUITE

	icvlen LENGTH
		length of the Integrity Check Value (ICV).

	[no-]encrypt
		authenticated encryption, or authenticity only

	[no-]send_sci
		include SCI in every packet, or only when it is necessary

	[no-]end_station
		End Station bit

	[no-]scb
		Single Copy Broadcast bit

	[no-]protect
		device MACsec protection

	[no-]replay
		device replay protection

	window SIZE
		replay window size

	validate { strict | check | disabled }

	encodingsa AN
		active secure association for transmission

TYPE VRF
	ip link add DEVICE type vrf table TABLE

	table TABLE
		id associated with VRF device

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

EXAMPLES
	ip link set dev ppp0 mtu 1400
		Change the MTU the ppp0 device.

	ip link add link eth0 name eth0.10 type vlan id 10
		Creates a new vlan device eth0.10 on device eth0.

	ip link delete dev eth0.10
		 Removes vlan device.

	ip link help gre
		Display help for the gre link type.

	ip link add name tun1 type ipip remote 192.168.1.1 local 192.168.1.2 \
		ttl 225 encap gue encap-sport auto encap-dport 5555 \
		encap-csum encap-remcsum

		Creates an IPIP that is encapsulated with Generic UDP
		Encapsula‐ tion, and the outer UDP checksum and remote checksum
		offload are enabled.

	ip link add link wpan0 lowpan0 type lowpan
	    Creates a 6LoWPAN interface named lowpan0 on the underlying IEEE
	    802.15.4 device wpan0.

SEE ALSO
	ip man link || ip link -man
	man ip || ip -man
	ip man netns || ip netns -man
	ethtool(8), iptables(8)

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
