// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
)

const (
	// Types < 0x600 (1536) are LLC packet lengths.
	TYPE_LLC_LENGTH              Type = 0x600
	TYPE_XNS_IDP                 Type = 0x600
	TYPE_IP4                     Type = 0x800
	TYPE_ARP                     Type = 0x806
	TYPE_VINES_IP                Type = 0x0BAD
	TYPE_VINES_LOOPBACK          Type = 0x0BAE
	TYPE_VINES_ECHO              Type = 0x0BAF
	TYPE_TRAIN                   Type = 0x1984
	TYPE_CDP                     Type = 0x2000
	TYPE_CGMP                    Type = 0x2001
	TYPE_SRP_CONTROL             Type = 0x2007
	TYPE_CENTRINO_PROMISC        Type = 0x2452
	TYPE_DECNET                  Type = 0x6000
	TYPE_DECNET_DUMP_LOAD        Type = 0x6001
	TYPE_DECNET_REMOTE_CONSOLE   Type = 0x6002
	TYPE_DECNET_ROUTE            Type = 0x6003
	TYPE_DEC_LAT                 Type = 0x6004
	TYPE_DEC_DIAGNOSTIC          Type = 0x6005
	TYPE_DEC_CUSTOMER            Type = 0x6006
	TYPE_DEC_SCA                 Type = 0x6007
	TYPE_TRANSPARENT_BRIDGING    Type = 0x6558
	TYPE_RAW_FRAME_RELAY         Type = 0x6559
	TYPE_REVERSE_ARP             Type = 0x8035
	TYPE_DEC_LAN_BRIDGE          Type = 0x8038
	TYPE_DEC_ETHERNET_ENCRYPTION Type = 0x803D
	TYPE_DEC_LAN_TRAFFIC_MONITOR Type = 0x803F
	TYPE_DEC_LAST                Type = 0x8041
	TYPE_APPLETALK               Type = 0x809B
	TYPE_IBM_SNA                 Type = 0x80D5
	TYPE_APPLETALK_AARP          Type = 0x80F3
	TYPE_WELLFLEET_COMPRESSION   Type = 0x80FF
	TYPE_VLAN                    Type = 0x8100
	TYPE_VLAN_IN_VLAN            Type = 0x9100
	TYPE_VLAN_802_1AD            Type = 0x88a8
	TYPE_IPX                     Type = 0x8137
	TYPE_SNMP                    Type = 0x814C
	TYPE_CABLETRON_ISMP          Type = 0x81FD
	TYPE_CABLETRON_ISMP_TBFLOOD  Type = 0x81FF
	TYPE_IP6                     Type = 0x86DD
	TYPE_ATOMIC                  Type = 0x86DF
	TYPE_TCP_IP_COMPRESSION      Type = 0x876B
	TYPE_IP_AUTONOMOUS_SYSTEMS   Type = 0x876C
	TYPE_SECURE_DATA             Type = 0x876D
	TYPE_MAC_CONTROL             Type = 0x8808
	TYPE_SLOW_PROTOCOLS          Type = 0x8809
	TYPE_PPP                     Type = 0x880B
	TYPE_MPLS_UNICAST            Type = 0x8847
	TYPE_MPLS_MULTICAST          Type = 0x8848
	TYPE_PPPOE_DISCOVERY         Type = 0x8863
	TYPE_PPPOE_SESSION           Type = 0x8864
	TYPE_INTEL_ANS               Type = 0x886D
	TYPE_MICROSOFT_NLB_HEARTBEAT Type = 0x886F
	TYPE_CDMA_2000               Type = 0x8881
	TYPE_PROFINET                Type = 0x8892
	TYPE_HYPERSCSI               Type = 0x889a
	TYPE_AOE                     Type = 0x88a2
	TYPE_BRDWALK                 Type = 0x88AE
	TYPE_LOOPBACK                Type = 0x9000
	TYPE_RTNET_MAC               Type = 0x9021
	TYPE_RTNET_CONFIG            Type = 0x9022
	TYPE_PGLAN                   Type = 0x9999
	TYPE_SRP_ISIS                Type = 0xFEFE
	TYPE_RESERVED                Type = 0xFFFF
)

var typeStrings = [...]string{
	TYPE_XNS_IDP:                 "XNS_IDP",
	TYPE_IP4:                     "IP4",
	TYPE_ARP:                     "ARP",
	TYPE_VINES_IP:                "VINES_IP",
	TYPE_VINES_LOOPBACK:          "VINES_LOOPBACK",
	TYPE_VINES_ECHO:              "VINES_ECHO",
	TYPE_TRAIN:                   "TRAIN",
	TYPE_CDP:                     "CDP",
	TYPE_CGMP:                    "CGMP",
	TYPE_SRP_CONTROL:             "SRP_CONTROL",
	TYPE_CENTRINO_PROMISC:        "CENTRINO_PROMISC",
	TYPE_DECNET:                  "DECNET",
	TYPE_DECNET_DUMP_LOAD:        "DECNET_DUMP_LOAD",
	TYPE_DECNET_REMOTE_CONSOLE:   "DECNET_REMOTE_CONSOLE",
	TYPE_DECNET_ROUTE:            "DECNET_ROUTE",
	TYPE_DEC_LAT:                 "DEC_LAT",
	TYPE_DEC_DIAGNOSTIC:          "DEC_DIAGNOSTIC",
	TYPE_DEC_CUSTOMER:            "DEC_CUSTOMER",
	TYPE_DEC_SCA:                 "DEC_SCA",
	TYPE_TRANSPARENT_BRIDGING:    "TRANSPARENT_BRIDGING",
	TYPE_RAW_FRAME_RELAY:         "RAW_FRAME_RELAY",
	TYPE_REVERSE_ARP:             "REVERSE_ARP",
	TYPE_DEC_LAN_BRIDGE:          "DEC_LAN_BRIDGE",
	TYPE_DEC_ETHERNET_ENCRYPTION: "DEC_ETHERNET_ENCRYPTION",
	TYPE_DEC_LAN_TRAFFIC_MONITOR: "DEC_LAN_TRAFFIC_MONITOR",
	TYPE_DEC_LAST:                "DEC_LAST",
	TYPE_APPLETALK:               "APPLETALK",
	TYPE_IBM_SNA:                 "IBM_SNA",
	TYPE_APPLETALK_AARP:          "APPLETALK_AARP",
	TYPE_WELLFLEET_COMPRESSION:   "WELLFLEET_COMPRESSION",
	TYPE_VLAN:                    "VLAN",
	TYPE_VLAN_IN_VLAN:            "VLAN_IN_VLAN",
	TYPE_VLAN_802_1AD:            "VLAN 802.1ad",
	TYPE_IPX:                     "IPX",
	TYPE_SNMP:                    "SNMP",
	TYPE_CABLETRON_ISMP:          "CABLETRON_ISMP",
	TYPE_CABLETRON_ISMP_TBFLOOD:  "CABLETRON_ISMP_TBFLOOD",
	TYPE_IP6:                     "IP6",
	TYPE_ATOMIC:                  "ATOMIC",
	TYPE_TCP_IP_COMPRESSION:      "TCP_IP_COMPRESSION",
	TYPE_IP_AUTONOMOUS_SYSTEMS:   "IP_AUTONOMOUS_SYSTEMS",
	TYPE_SECURE_DATA:             "SECURE_DATA",
	TYPE_MAC_CONTROL:             "MAC_CONTROL",
	TYPE_SLOW_PROTOCOLS:          "SLOW_PROTOCOLS",
	TYPE_PPP:                     "PPP",
	TYPE_MPLS_UNICAST:            "MPLS_UNICAST",
	TYPE_MPLS_MULTICAST:          "MPLS_MULTICAST",
	TYPE_PPPOE_DISCOVERY:         "PPPOE_DISCOVERY",
	TYPE_PPPOE_SESSION:           "PPPOE_SESSION",
	TYPE_INTEL_ANS:               "INTEL_ANS",
	TYPE_MICROSOFT_NLB_HEARTBEAT: "MICROSOFT_NLB_HEARTBEAT",
	TYPE_CDMA_2000:               "CDMA_2000",
	TYPE_PROFINET:                "PROFINET",
	TYPE_HYPERSCSI:               "HYPERSCSI",
	TYPE_AOE:                     "AOE",
	TYPE_BRDWALK:                 "BRDWALK",
	TYPE_LOOPBACK:                "LOOPBACK",
	TYPE_RTNET_MAC:               "RTNET_MAC",
	TYPE_RTNET_CONFIG:            "RTNET_CONFIG",
	TYPE_PGLAN:                   "PGLAN",
	TYPE_SRP_ISIS:                "SRP_ISIS",
	TYPE_RESERVED:                "RESERVED",
}

func (t Type) String() string {
	return elib.StringerHex(typeStrings[:], int(t))
}

var typeMap = parse.NewStringMap(typeStrings[:])

func (t *Type) Parse(in *parse.Input) {
	var v uint16
	switch {
	case in.Parse("%v", typeMap, &v):
	case in.Parse("%v", &v):
	default:
		panic(parse.ErrInput)
	}
	*t = Type(v).FromHost()
}
