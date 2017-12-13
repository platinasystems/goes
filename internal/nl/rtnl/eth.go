// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	ETH_P_LOOP      uint16 = 0x0060 // Ethernet Loopback packet
	ETH_P_PUP       uint16 = 0x0200 // Xerox PUP packet
	ETH_P_PUPAT     uint16 = 0x0201 // Xerox PUP Addr Trans packet
	ETH_P_TSN       uint16 = 0x22F0 // TSN (IEEE 1722) packet
	ETH_P_IP        uint16 = 0x0800 // Internet Protocol packet
	ETH_P_X25       uint16 = 0x0805 // CCITT X.25
	ETH_P_ARP       uint16 = 0x0806 // Address Resolution packet
	ETH_P_BPQ       uint16 = 0x08FF // G8BPQ AX.25 Ethernet Packet	[ NOT AN OFFICIALLY REGISTERED ID ]
	ETH_P_IEEEPUP   uint16 = 0x0a00 // Xerox IEEE802.3 PUP packet
	ETH_P_IEEEPUPAT uint16 = 0x0a01 // Xerox IEEE802.3 PUP Addr Trans packet
	ETH_P_BATMAN    uint16 = 0x4305 // B.A.T.M.A.N.-Advanced packet [ NOT AN OFFICIALLY REGISTERED ID ]
	ETH_P_DEC       uint16 = 0x6000 // DEC Assigned proto
	ETH_P_DNA_DL    uint16 = 0x6001 // DEC DNA Dump/Load
	ETH_P_DNA_RC    uint16 = 0x6002 // DEC DNA Remote Console
	ETH_P_DNA_RT    uint16 = 0x6003 // DEC DNA Routing
	ETH_P_LAT       uint16 = 0x6004 // DEC LAT
	ETH_P_DIAG      uint16 = 0x6005 // DEC Diagnostics
	ETH_P_CUST      uint16 = 0x6006 // DEC Customer use
	ETH_P_SCA       uint16 = 0x6007 // DEC Systems Comms Arch
	ETH_P_TEB       uint16 = 0x6558 // Trans Ether Bridging
	ETH_P_RARP      uint16 = 0x8035 // Reverse Addr Res packet
	ETH_P_ATALK     uint16 = 0x809B // Appletalk DDP
	ETH_P_AARP      uint16 = 0x80F3 // Appletalk AARP
	ETH_P_8021Q     uint16 = 0x8100 // 802.1Q VLAN Extended Header
	ETH_P_IPX       uint16 = 0x8137 // IPX over DIX
	ETH_P_IPV6      uint16 = 0x86DD // IPv6 over bluebook
	ETH_P_PAUSE     uint16 = 0x8808 // IEEE Pause frames. See 802.3 31B
	ETH_P_SLOW      uint16 = 0x8809 // Slow Protocol. See 802.3ad 43B
	ETH_P_WCCP      uint16 = 0x883E // Web-cache coordination
	ETH_P_MPLS_UC   uint16 = 0x8847 // MPLS Unicast traffic
	ETH_P_MPLS_MC   uint16 = 0x8848 // MPLS Multicast traffic
	ETH_P_ATMMPOA   uint16 = 0x884c // MultiProtocol Over ATM
	ETH_P_PPP_DISC  uint16 = 0x8863 // PPPoE discovery messages
	ETH_P_PPP_SES   uint16 = 0x8864 // PPPoE session messages
	ETH_P_LINK_CTL  uint16 = 0x886c // HPNA, wlan link local tunnel
	ETH_P_ATMFATE   uint16 = 0x8884 // Frame-based ATM Transport
	// over Ethernet
	ETH_P_PAE     uint16 = 0x888E // Port Access Entity (IEEE 802.1X)
	ETH_P_AOE     uint16 = 0x88A2 // ATA over Ethernet
	ETH_P_8021AD  uint16 = 0x88A8 // 802.1ad Service VLAN
	ETH_P_802_EX1 uint16 = 0x88B5 // 802.1 Local Experimental 1.
	ETH_P_TIPC    uint16 = 0x88CA // TIPC
	ETH_P_MACSEC  uint16 = 0x88E5 // 802.1ae MACsec
	ETH_P_8021AH  uint16 = 0x88E7 // 802.1ah Backbone Service Tag
	ETH_P_MVRP    uint16 = 0x88F5 // 802.1Q MVRP
	ETH_P_1588    uint16 = 0x88F7 // IEEE 1588 Timesync
	ETH_P_NCSI    uint16 = 0x88F8 // NCSI protocol
	ETH_P_PRP     uint16 = 0x88FB // IEC 62439-3 PRP/HSRv0
	ETH_P_FCOE    uint16 = 0x8906 // Fibre Channel over Ethernet
	ETH_P_IBOE    uint16 = 0x8915 // Infiniband over Ethernet
	ETH_P_TDLS    uint16 = 0x890D // TDLS
	ETH_P_FIP     uint16 = 0x8914 // FCoE Initialization Protocol
	ETH_P_80221   uint16 = 0x8917 // IEEE 802.21
	// Media Independent Handover Protocol
	ETH_P_HSR      uint16 = 0x892F // IEC 62439-3 HSRv1
	ETH_P_LOOPBACK uint16 = 0x9000 // IEEE 802.3 Ethernet loopback

	// unofficial ids...
	ETH_P_QINQ1   uint16 = 0x9100 // deprecated QinQ VLAN
	ETH_P_QINQ2   uint16 = 0x9200 // deprecated QinQ VLAN
	ETH_P_QINQ3   uint16 = 0x9300 // deprecated QinQ VLAN
	ETH_P_EDSA    uint16 = 0xDADA // Ethertype DSA
	ETH_P_AF_IUCV uint16 = 0xFBFB // IBM af_iucv

	// If the value in the ethernet type is less than this value
	// then the frame is Ethernet II. Else it is 802.3
	ETH_P_802_3_MIN uint16 = 0x0600

	// Non DIX types. Won't clash for 1500 types.
	ETH_P_802_3      uint16 = 0x0001 // Dummy type for 802.3 frames
	ETH_P_AX25       uint16 = 0x0002 // Dummy protocol id for AX.25
	ETH_P_ALL        uint16 = 0x0003 // Every packet (be careful!!!)
	ETH_P_802_2      uint16 = 0x0004 // 802.2 frames
	ETH_P_SNAP       uint16 = 0x0005 // Internal only
	ETH_P_DDCMP      uint16 = 0x0006 // DEC DDCMP: Internal only
	ETH_P_WAN_PPP    uint16 = 0x0007 // Dummy type for WAN PPP frames*/
	ETH_P_PPP_MP     uint16 = 0x0008 // Dummy type for PPP MP frames
	ETH_P_LOCALTALK  uint16 = 0x0009 // Localtalk pseudo type
	ETH_P_CAN        uint16 = 0x000C // CAN: Controller Area Network
	ETH_P_CANFD      uint16 = 0x000D // CANFD: CAN flexible data rate*/
	ETH_P_PPPTALK    uint16 = 0x0010 // Dummy type for Atalk over PPP*/
	ETH_P_TR_802_2   uint16 = 0x0011 // 802.2 frames
	ETH_P_MOBITEX    uint16 = 0x0015 // Mobitex (kaz@cafe.net)
	ETH_P_CONTROL    uint16 = 0x0016 // Card specific control frames
	ETH_P_IRDA       uint16 = 0x0017 // Linux-IrDA
	ETH_P_ECONET     uint16 = 0x0018 // Acorn Econet
	ETH_P_HDLC       uint16 = 0x0019 // HDLC frames
	ETH_P_ARCNET     uint16 = 0x001A // 1A for ArcNet :-)
	ETH_P_DSA        uint16 = 0x001B // Distributed Switch Arch.
	ETH_P_TRAILER    uint16 = 0x001C // Trailer switch tagging
	ETH_P_PHONET     uint16 = 0x00F5 // Nokia Phonet frames
	ETH_P_IEEE802154 uint16 = 0x00F6 // IEEE802.15.4 frame
	ETH_P_CAIF       uint16 = 0x00F7 // ST-Ericsson CAIF protocol
	ETH_P_XDSA       uint16 = 0x00F8 // Multiplexed DSA protocol
)
