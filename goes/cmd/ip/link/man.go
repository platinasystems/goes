// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

const Man = `
DESCRIPTION
   ip link add - add virtual link
	link DEVICE
	       specifies the physical device to act operate on.

	       NAME specifies the name of the new virtual device.

	       TYPE specifies the type of the new device.

	       Link types:

	               bridge - Ethernet Bridge device

	               bond - Bonding device can - Controller Area Network
	               interface

	               dummy - Dummy network interface

	               hsr - High-availability Seamless Redundancy device

	               ifb - Intermediate Functional Block device

	               ipoib - IP over Infiniband device

	               macvlan - Virtual interface base on link layer address
	               (MAC)

	               macvtap - Virtual interface based on link layer address
	               (MAC) and TAP.

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

	               ipvlan - Interface for L3 (IPv6/IPv4) based VLANs

	               lowpan - Interface for 6LoWPAN (IPv6) over IEEE 802.15.4
	               / Bluetooth

	               geneve - GEneric NEtwork Virtualization Encapsulation

	               macsec - Interface for IEEE 802.1AE MAC Security (MAC‐
	               sec)

	               vrf - Interface for L3 VRF domains

	numtxqueues QUEUE_COUNT
	       specifies the number of transmit queues for new device.

	numrxqueues QUEUE_COUNT
	       specifies the number of receive queues for new device.

	index IDX
	       specifies the desired index of the new virtual device. The link
	       creation fails, if the index is busy.

	VLAN Type Support
	       For a link of type VLAN the following additional arguments are
	       supported:

	       ip link add link DEVICE name NAME type vlan [ protocol
	       VLAN_PROTO ] id VLANID [ reorder_hdr { on | off } ] [ gvrp { on
	       | off } ] [ mvrp { on | off } ] [ loose_binding { on | off } ] [
	       ingress-qos-map QOS-MAP ] [ egress-qos-map QOS-MAP ]

	               protocol VLAN_PROTO - either 802.1Q or 802.1ad.

	               id VLANID - specifies the VLAN Identifer to use. Note
	               that numbers with a leading " 0 " or " 0x " are inter‐
	               preted as octal or hexadeimal, respectively.

	               reorder_hdr { on | off } - specifies whether ethernet
	               headers are reordered or not (default is on).

	                   If reorder_hdr is on then VLAN header will be not
	                   inserted immediately but only before passing to the
	                   physical device (if this device does not support
	                   VLAN offloading), the similar on the RX direction -
	                   by default the packet will be untagged before being
	                   received by VLAN device. Reordering allows to accel‐
	                   erate tagging on egress and to hide VLAN header on
	                   ingress so the packet looks like regular Ethernet
	                   packet, at the same time it might be confusing for
	                   packet capture as the VLAN header does not exist
	                   within the packet.

	                   VLAN offloading can be checked by ethtool(8):

	                       ethtool -k <phy_dev> | grep tx-vlan-offload

	                   where <phy_dev> is the physical device to which VLAN
	                   device is bound.

	               gvrp { on | off } - specifies whether this VLAN should
	               be registered using GARP VLAN Registration Protocol.

	               mvrp { on | off } - specifies whether this VLAN should
	               be registered using Multiple VLAN Registration Protocol.

	               loose_binding { on | off } - specifies whether the VLAN
	               device state is bound to the physical device state.

	               ingress-qos-map QOS-MAP - defines a mapping of VLAN
	               header prio field to the Linux internal packet priority
	               on incoming frames. The format is FROM:TO with multiple
	               mappings separated by spaces.

	               egress-qos-map QOS-MAP - defines a mapping of Linux
	               internal packet priority to VLAN header prio field but
	               for outgoing frames. The format is the same as for
	               ingress-qos-map.

	                   Linux packet priority can be set by iptables(8):

	                       iptables -t mangle -A POSTROUTING [...] -j CLAS‐
	                       SIFY --set-class 0:4

	                   and this "4" priority can be used in the egress qos
	                   mapping to set VLAN prio "5":

	                       ip link set veth0.10 type vlan egress 4:5

	VXLAN Type Support
	       For a link of type VXLAN the following additional arguments are
	       supported:

	       ip link add DEVICE type vxlan id VNI [ dev PHYS_DEV  ] [ { group
	       | remote } IPADDR ] [ local { IPADDR | any } ] [ ttl TTL ] [ tos
	       TOS ] [ flowlabel FLOWLABEL ] [ dstport PORT ] [ srcport MIN MAX
	       ] [ [no]learning ] [ [no]proxy ] [ [no]rsc ] [ [no]l2miss ] [
	       [no]l3miss ] [ [no]udpcsum ] [ [no]udp6zerocsumtx ] [
	       [no]udp6zerocsumrx ] [ ageing SECONDS ] [ maxaddress NUMBER ] [
	       [no]external ] [ gbp ] [ gpe ]

	               id VNI - specifies the VXLAN Network Identifer (or VXLAN
	               Segment Identifier) to use.

	               dev PHYS_DEV - specifies the physical device to use for
	               tunnel endpoint communication.

	               group IPADDR - specifies the multicast IP address to
	               join.  This parameter cannot be specified with the
	               remote parameter.

	               remote IPADDR - specifies the unicast destination IP
	               address to use in outgoing packets when the destination
	               link layer address is not known in the VXLAN device for‐
	               warding database. This parameter cannot be specified
	               with the group parameter.

	               local IPADDR - specifies the source IP address to use in
	               outgoing packets.

	               ttl TTL - specifies the TTL value to use in outgoing
	               packets.

	               tos TOS - specifies the TOS value to use in outgoing
	               packets.

	               flowlabel FLOWLABEL - specifies the flow label to use in
	               outgoing packets.

	               dstport PORT - specifies the UDP destination port to
	               communicate to the remote VXLAN tunnel endpoint.

	               srcport MIN MAX - specifies the range of port numbers to
	               use as UDP source ports to communicate to the remote
	               VXLAN tunnel endpoint.

	               [no]learning - specifies if unknown source link layer
	               addresses and IP addresses are entered into the VXLAN
	               device forwarding database.

	               [no]rsc - specifies if route short circuit is turned on.

	               [no]proxy - specifies ARP proxy is turned on.

	               [no]l2miss - specifies if netlink LLADDR miss notifica‐
	               tions are generated.

	               [no]l3miss - specifies if netlink IP ADDR miss notifica‐
	               tions are generated.

	               [no]udpcsum - specifies if UDP checksum is calculated
	               for transmitted packets over IPv4.

	               [no]udp6zerocsumtx - skip UDP checksum calculation for
	               transmitted packets over IPv6.

	               [no]udp6zerocsumrx - allow incoming UDP packets over
	               IPv6 with zero checksum field.

	               ageing SECONDS - specifies the lifetime in seconds of
	               FDB entries learnt by the kernel.

	               maxaddress NUMBER - specifies the maximum number of FDB
	               entries.

	               [no]external - specifies whether an external control
	               plane (e.g. ip route encap) or the internal FDB should
	               be used.

	               gbp - enables the Group Policy extension (VXLAN-GBP).

	                   Allows to transport group policy context across
	                   VXLAN network peers.  If enabled, includes the mark
	                   of a packet in the VXLAN header for outgoing packets
	                   and fills the packet mark based on the information
	                   found in the VXLAN header for incomming packets.

	                   Format of upper 16 bits of packet mark (flags);

	                     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	                     |-|-|-|-|-|-|-|-|-|D|-|-|A|-|-|-|
	                     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	                     D := Don't Learn bit. When set, this bit indicates
	                     that the egress VTEP MUST NOT learn the source
	                     address of the encapsulated frame.

	                     A := Indicates that the group policy has already
	                     been applied to this packet. Policies MUST NOT be
	                     applied by devices when the A bit is set.

	                   Format of lower 16 bits of packet mark (policy ID):

	                     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	                     |        Group Policy ID        |
	                     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	                   Example:
	                     iptables -A OUTPUT [...] -j MARK --set-mark
	                   0x800FF

	               gpe - enables the Generic Protocol extension (VXLAN-
	               GPE). Currently, this is only supported together with
	               the external keyword.

	GRE, IPIP, SIT Type Support
	       For a link of types GRE/IPIP/SIT the following additional argu‐
	       ments are supported:

	       ip link add DEVICE type { gre | ipip | sit }  remote ADDR local
	       ADDR [ encap { fou | gue | none } ] [ encap-sport { PORT | auto
	       } ] [ encap-dport PORT ] [ [no]encap-csum ] [ [no]encap-remcsum
	       ]

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

	               [no]encap-csum - specifies if UDP checksums are enabled
	               in the secondary encapsulation.

	               [no]encap-remcsum - specifies if Remote Checksum Offload
	               is enabled. This is only applicable for Generic UDP
	               Encapsulation.

	IP6GRE/IP6GRETAP Type Support
	       For a link of type IP6GRE/IP6GRETAP the following additional
	       arguments are supported:

	       ip link add DEVICE type { ip6gre | ip6gretap } remote ADDR local
	       ADDR [ [i|o]seq ] [ [i|o]key KEY ] [ [i|o]csum ] [ hoplimit TTL
	       ] [ encaplimit ELIM ] [ tclass TCLASS ] [ flowlabel FLOWLABEL ]
	       [ dscp inherit ] [ dev PHYS_DEV ]

	               remote ADDR - specifies the remote IPv6 address of the
	               tunnel.

	               local ADDR - specifies the fixed local IPv6 address for
	               tunneled packets.  It must be an address on another
	               interface on this host.

	               [i|o]seq - serialize packets.  The oseq flag enables
	               sequencing of outgoing packets.  The iseq flag requires
	               that all input packets are serialized.

	               [i|o]key KEY - use keyed GRE with key KEY. KEY is either
	               a number or an IPv4 address-like dotted quad.  The key
	               parameter specifies the same key to use in both direc‐
	               tions.  The ikey and okey parameters specify different
	               keys for input and output.

	               [i|o]csum - generate/require checksums for tunneled
	               packets.  The ocsum flag calculates checksums for outgo‐
	               ing packets.  The icsum flag requires that all input
	               packets have the correct checksum. The csum flag is
	               equivalent to the combination icsum ocsum.

	               hoplimit TTL - specifies Hop Limit value to use in out‐
	               going packets.

	               encaplimit ELIM - specifies a fixed encapsulation limit.
	               Default is 4.

	               flowlabel FLOWLABEL - specifies a fixed flowlabel.

	               tclass TCLASS - specifies the traffic class field on
	               tunneled packets, which can be specified as either a
	               two-digit hex value (e.g. c0) or a predefined string
	               (e.g. internet).  The value inherit causes the field to
	               be copied from the original IP header. The values
	               inherit/STRING or inherit/00..ff will set the field to
	               STRING or 00..ff when tunneling non-IP packets. The
	               default value is 00.

	IPoIB Type Support
	       For a link of type IPoIB the following additional arguments are
	       supported:

	       ip link add DEVICE name NAME type ipoib [ pkey PKEY ] [ mode
	       MODE ]

	               pkey PKEY - specifies the IB P-Key to use.

	               mode MODE - specifies the mode (datagram or connected)
	               to use.

	GENEVE Type Support
	       For a link of type GENEVE the following additional arguments are
	       supported:

	       ip link add DEVICE type geneve id VNI remote IPADDR [ ttl TTL ]
	       [ tos TOS ] [ flowlabel FLOWLABEL ]

	               id VNI - specifies the Virtual Network Identifer to use.

	               remote IPADDR - specifies the unicast destination IP
	               address to use in outgoing packets.

	               ttl TTL - specifies the TTL value to use in outgoing
	               packets.

	               tos TOS - specifies the TOS value to use in outgoing
	               packets.

	               flowlabel FLOWLABEL - specifies the flow label to use in
	               outgoing packets.

	MACVLAN and MACVTAP Type Support
	       For a link of type MACVLAN or MACVTAP the following additional
	       arguments are supported:

	       ip link add link DEVICE name NAME type { macvlan | macvtap }
	       mode { private | vepa | bridge | passthru  [ nopromisc ] }

	               type { macvlan | macvtap } - specifies the link type to
	               use.  macvlan creates just a virtual interface, while
	               macvtap in addition creates a character device /dev/tapX
	               to be used just like a tuntap device.

	               mode private - Do not allow communication between
	               macvlan instances on the same physical interface, even
	               if the external switch supports hairpin mode.

	               mode vepa - Virtual Ethernet Port Aggregator mode. Data
	               from one macvlan instance to the other on the same phys‐
	               ical interface is transmitted over the physical inter‐
	               face. Either the attached switch needs to support hair‐
	               pin mode, or there must be a TCP/IP router forwarding
	               the packets in order to allow communication. This is the
	               default mode.

	               mode bridge - In bridge mode, all endpoints are directly
	               connected to each other, communication is not redirected
	               through the physical interface's peer.

	               mode passthru [ nopromisc ] - This mode gives more power
	               to a single endpoint, usually in macvtap mode. It is not
	               allowed for more than one endpoint on the same physical
	               interface. All traffic will be forwarded to this end‐
	               point, allowing virtio guests to change MAC address or
	               set promiscuous mode in order to bridge the interface or
	               create vlan interfaces on top of it. By default, this
	               mode forces the underlying interface into promiscuous
	               mode. Passing the nopromisc flag prevents this, so the
	               promisc flag may be controlled using standard tools.

	High-availability Seamless Redundancy (HSR) Support
	       For a link of type HSR the following additional arguments are
	       supported:

	       ip link add link DEVICE name NAME type hsr slave1 SLAVE1-IF
	       slave2 SLAVE2-IF [ supervision ADDR-BYTE ] [ version { 0 | 1 } ]

	               type hsr - specifies the link type to use, here HSR.

	               slave1 SLAVE1-IF - Specifies the physical device used
	               for the first of the two ring ports.

	               slave2 SLAVE2-IF - Specifies the physical device used
	               for the second of the two ring ports.

	               supervision ADDR-BYTE - The last byte of the multicast
	               address used for HSR supervision frames.  Default option
	               is "0", possible values 0-255.

	               version { 0 | 1 } - Selects the protocol version of the
	               interface. Default option is "0", which corresponds to
	               the 2010 version of the HSR standard. Option "1" acti‐
	               vates the 2012 version.

	MACsec Type Support
	       For a link of type MACsec the following additional arguments are
	       supported:

	       ip link add link DEVICE name NAME type macsec [ port PORT | sci
	       SCI ] [ cipher CIPHER_SUITE ] [ icvlen { 8..16 } ] [ encrypt {
	       on | off } ] [ send_sci { on | off } ] [ end_station { on | off
	       } ] [ scb { on | off } ] [ protect { on | off } ] [ replay { on
	       | off } window { 0..2^32-1 } ] [ validate { strict | check |
	       disabled } ] [ encodingsa { 0..3 } ]

	               port PORT - sets the port number for this MACsec device.

	               sci SCI - sets the SCI for this MACsec device.

	               cipher CIPHER_SUITE - defines the cipher suite to use.

	               icvlen LENGTH - sets the length of the Integrity Check
	               Value (ICV).

	               encrypt on or encrypt off - switches between authenti‐
	               cated encryption, or authenticity mode only.

	               send_sci on or send_sci off - specifies whether the SCI
	               is included in every packet, or only when it is neces‐
	               sary.

	               end_station on or end_station off - sets the End Station
	               bit.

	               scb on or scb off - sets the Single Copy Broadcast bit.

	               protect on or protect off - enables MACsec protection on
	               the device.

	               replay on or replay off - enables replay protection on
	               the device.

	                       window SIZE - sets the size of the replay win‐
	                       dow.

	               validate strict or validate check or validate disabled -
	               sets the validation mode on the device.

	               encodingsa AN - sets the active secure association for
	               transmission.

	VRF Type Support
	       For a link of type VRF the following additional arguments are
	       supported:

	       ip link add DEVICE type vrf table TABLE

	               table table id associated with VRF device

   ip link delete - delete virtual link
	dev DEVICE
	       specifies the virtual device to act operate on.

	group GROUP
	       specifies the group of virtual links to delete. Group 0 is not
	       allowed to be deleted since it is the default group.

	type TYPE
	       specifies the type of the device.

   ip link set - change device attributes
	Warning: If multiple parameter changes are requested, ip aborts immedi‐
	ately after any of the changes have failed.  This is the only case when
	ip can move the system to an unpredictable state. The solution is to
	avoid changing several parameters with one ip link set call.

	dev DEVICE
	       DEVICE specifies network device to operate on. When configuring
	       SR-IOV Virtual Function (VF) devices, this keyword should spec‐
	       ify the associated Physical Function (PF) device.

	group GROUP
	       GROUP has a dual role: If both group and dev are present, then
	       move the device to the specified group. If only a group is spec‐
	       ified, then the command operates on all devices in that group.

	up and down
	       change the state of the device to UP or DOWN.

	arp on or arp off
	       change the NOARP flag on the device.

	multicast on or multicast off
	       change the MULTICAST flag on the device.

	protodown on or protodown off
	       change the PROTODOWN state on the device. Indicates that a pro‐
	       tocol error has been detected on the port. Switch drivers can
	       react to this error by doing a phys down on the switch port.

	dynamic on or dynamic off
	       change the DYNAMIC flag on the device. Indicates that address
	       can change when interface goes down (currently NOT used by the
	       Linux).

	name NAME
	       change the name of the device. This operation is not recommended
	       if the device is running or has some addresses already config‐
	       ured.

	txqueuelen NUMBER

	txqlen NUMBER
	       change the transmit queue length of the device.

	mtu NUMBER
	       change the MTU of the device.

	address LLADDRESS
	       change the station address of the interface.

	broadcast LLADDRESS

	brd LLADDRESS

	peer LLADDRESS
	       change the link layer broadcast address or the peer address when
	       the interface is POINTOPOINT.

	netns NETNSNAME | PID
	       move the device to the network namespace associated with name
	       NETNSNAME or process PID.

	       Some devices are not allowed to change network namespace: loop‐
	       back, bridge, ppp, wireless. These are network namespace local
	       devices. In such case ip tool will return "Invalid argument"
	       error. It is possible to find out if device is local to a single
	       network namespace by checking netns-local flag in the output of
	       the ethtool:

	               ethtool -k DEVICE

	       To change network namespace for wireless devices the iw tool can
	       be used. But it allows to change network namespace only for
	       physical devices and by process PID.

	alias NAME
	       give the device a symbolic name for easy reference.

	group GROUP
	       specify the group the device belongs to.  The available groups
	       are listed in file /etc/iproute2/group.

	vf NUM specify a Virtual Function device to be configured. The associ‐
	       ated PF device must be specified using the dev parameter.

	               mac LLADDRESS - change the station address for the spec‐
	               ified VF. The vf parameter must be specified.

	               vlan VLANID - change the assigned VLAN for the specified
	               VF. When specified, all traffic sent from the VF will be
	               tagged with the specified VLAN ID. Incoming traffic will
	               be filtered for the specified VLAN ID, and will have all
	               VLAN tags stripped before being passed to the VF. Set‐
	               ting this parameter to 0 disables VLAN tagging and fil‐
	               tering. The vf parameter must be specified.

	               qos VLAN-QOS - assign VLAN QOS (priority) bits for the
	               VLAN tag. When specified, all VLAN tags transmitted by
	               the VF will include the specified priority bits in the
	               VLAN tag. If not specified, the value is assumed to be
	               0. Both the vf and vlan parameters must be specified.
	               Setting both vlan and qos as 0 disables VLAN tagging and
	               filtering for the VF.

	               rate TXRATE -- change the allowed transmit bandwidth, in
	               Mbps, for the specified VF.  Setting this parameter to 0
	               disables rate limiting.  vf parameter must be specified.
	               Please use new API max_tx_rate option instead.

	               max_tx_rate TXRATE - change the allowed maximum transmit
	               bandwidth, in Mbps, for the specified VF.  vf parameter
	               must be specified.

	               min_tx_rate TXRATE - change the allowed minimum transmit
	               bandwidth, in Mbps, for the specified VF.  Minimum
	               TXRATE should be always <= Maximum TXRATE.  vf parameter
	               must be specified.

	               spoofchk on|off - turn packet spoof checking on or off
	               for the specified VF.

	               query_rss on|off - toggle the ability of querying the
	               RSS configuration of a specific VF. VF RSS information
	               like RSS hash key may be considered sensitive on some
	               devices where this information is shared between VF and
	               PF and thus its querying may be prohibited by default.

	               state auto|enable|disable - set the virtual link state
	               as seen by the specified VF. Setting to auto means a
	               reflection of the PF link state, enable lets the VF to
	               communicate with other VFs on this host even if the PF
	               link state is down, disable causes the HW to drop any
	               packets sent by the VF.

	               trust on|off - trust the specified VF user. This enables
	               that VF user can set a specific feature which may impact
	               security and/or performance. (e.g. VF multicast promis‐
	               cuous mode)

	               node_guid eui64 - configure node GUID for the VF.

	               port_guid eui64 - configure port GUID for the VF.

	master DEVICE
	       set master device of the device (enslave device).

	nomaster
	       unset master device of the device (release device).

	addrgenmode eui64|none|stable_secret|random
	       set the IPv6 address generation mode

	       eui64 - use a Modified EUI-64 format interface identifier

	       none - disable automatic address generation

	       stable_secret - generate the interface identifier based on a
	       preset /proc/sys/net/ipv6/conf/{default,DEVICE}/stable_secret

	       random - like stable_secret, but auto-generate a new random
	       secret if none is set

	link-netnsid
	       set peer netnsid for a cross-netns interface

	type ETYPE TYPE_ARGS
	       Change type-specific settings. For a list of supported types and
	       arguments refer to the description of ip link add above. In
	       addition to that, it is possible to manipulate settings to slave
	       devices:

	Bridge Slave Support
	       For a link with master bridge the following additional arguments
	       are supported:

	       ip link set type bridge_slave [ state STATE ] [ priority PRIO ]
	       [ cost COST ] [ guard { on | off } ] [ hairpin { on | off } ] [
	       fastleave { on | off } ] [ root_block { on | off } ] [ learning
	       { on | off } ] [ flood { on | off } ] [ proxy_arp { on | off } ]
	       [ proxy_arp_wifi { on | off } ] [ mcast_router MULTICAST_ROUTER
	       ] [ mcast_fast_leave { on | off} ]

	               state STATE - Set port state.  STATE is a number repre‐
	               senting the following states: 0 (disabled), 1 (listen‐
	               ing), 2 (learning), 3 (forwarding), 4 (blocking).

	               priority PRIO - set port priority (a 16bit unsigned
	               value).

	               cost COST - set port cost (a 32bit unsigned value).

	               guard { on | off } - block incoming BPDU packets on this
	               port.

	               hairpin { on | off } - enable hairpin mode on this port.
	               This will allow incoming packets on this port to be
	               reflected back.

	               fastleave { on | off } - enable multicast fast leave on
	               this port.

	               root_block { on | off } - block this port from becoming
	               the bridge's root port.

	               learning { on | off } - allow MAC address learning on
	               this port.

	               flood { on | off } - open the flood gates on this port,
	               i.e. forward all unicast frames to this port also.
	               Requires proxy_arp and proxy_arp_wifi to be turned off.

	               proxy_arp { on | off } - enable proxy ARP on this port.

	               proxy_arp_wifi { on | off } - enable proxy ARP on this
	               port which meets extended requirements by IEEE 802.11
	               and Hotspot 2.0 specifications.

	               mcast_router MULTICAST_ROUTER - configure this port for
	               having multicast routers attached. A port with a multi‐
	               cast router will receive all multicast traffic.  MULTI‐
	               CAST_ROUTER may be either 0 to disable multicast routers
	               on this port, 1 to let the system detect the presence of
	               of routers (this is the default), 2 to permanently
	               enable multicast traffic forwarding on this port or 3 to
	               enable multicast routers temporarily on this port, not
	               depending on incoming queries.

	               mcast_fast_leave { on | off } - this is a synonym to the
	               fastleave option above.

	Bonding Slave Support
	       For a link with master bond the following additional arguments
	       are supported:

	       ip link set type bond_slave [ queue_id ID ]

	               queue_id ID - set the slave's queue ID (a 16bit unsigned
	               value).

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

	ip link set dev ppp0 mtu 1400
	    Change the MTU the ppp0 device.

	ip link add link eth0 name eth0.10 type vlan id 10
	    Creates a new vlan device eth0.10 on device eth0.

	ip link delete dev eth0.10
	    Removes vlan device.

	ip link help gre
	    Display help for the gre link type.

	ip link add name tun1 type ipip remote 192.168.1.1 local 192.168.1.2
	ttl 225 encap gue encap-sport auto encap-dport 5555 encap-csum encap-
	remcsum
	    Creates an IPIP that is encapsulated with Generic UDP Encapsula‐
	    tion, and the outer UDP checksum and remote checksum offload are
	    enabled.

	ip link add link wpan0 lowpan0 type lowpan
	    Creates a 6LoWPAN interface named lowpan0 on the underlying IEEE
	    802.15.4 device wpan0.

SEE ALSO
	man ip || ip -man
	ip man netns || ip netns -man
	ethtool(8), iptables(8)

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
