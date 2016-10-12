package ethernet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
)

const (
	// Types < 0x600 (1536) are LLC packet lengths.
	LLC_LENGTH              Type = 0x600
	XNS_IDP                 Type = 0x600
	IP4                     Type = 0x800
	ARP                     Type = 0x806
	VINES_IP                Type = 0x0BAD
	VINES_LOOPBACK          Type = 0x0BAE
	VINES_ECHO              Type = 0x0BAF
	TRAIN                   Type = 0x1984
	CDP                     Type = 0x2000
	CGMP                    Type = 0x2001
	SRP_CONTROL             Type = 0x2007
	CENTRINO_PROMISC        Type = 0x2452
	DECNET                  Type = 0x6000
	DECNET_DUMP_LOAD        Type = 0x6001
	DECNET_REMOTE_CONSOLE   Type = 0x6002
	DECNET_ROUTE            Type = 0x6003
	DEC_LAT                 Type = 0x6004
	DEC_DIAGNOSTIC          Type = 0x6005
	DEC_CUSTOMER            Type = 0x6006
	DEC_SCA                 Type = 0x6007
	TRANSPARENT_BRIDGING    Type = 0x6558
	RAW_FRAME_RELAY         Type = 0x6559
	REVERSE_ARP             Type = 0x8035
	DEC_LAN_BRIDGE          Type = 0x8038
	DEC_ETHERNET_ENCRYPTION Type = 0x803D
	DEC_LAN_TRAFFIC_MONITOR Type = 0x803F
	DEC_LAST                Type = 0x8041
	APPLETALK               Type = 0x809B
	IBM_SNA                 Type = 0x80D5
	APPLETALK_AARP          Type = 0x80F3
	WELLFLEET_COMPRESSION   Type = 0x80FF
	VLAN                    Type = 0x8100
	VLAN_IN_VLAN            Type = 0x9100
	IPX                     Type = 0x8137
	SNMP                    Type = 0x814C
	CABLETRON_ISMP          Type = 0x81FD
	CABLETRON_ISMP_TBFLOOD  Type = 0x81FF
	IP6                     Type = 0x86DD
	ATOMIC                  Type = 0x86DF
	TCP_IP_COMPRESSION      Type = 0x876B
	IP_AUTONOMOUS_SYSTEMS   Type = 0x876C
	SECURE_DATA             Type = 0x876D
	MAC_CONTROL             Type = 0x8808
	SLOW_PROTOCOLS          Type = 0x8809
	PPP                     Type = 0x880B
	MPLS_UNICAST            Type = 0x8847
	MPLS_MULTICAST          Type = 0x8848
	PPPOE_DISCOVERY         Type = 0x8863
	PPPOE_SESSION           Type = 0x8864
	INTEL_ANS               Type = 0x886D
	MICROSOFT_NLB_HEARTBEAT Type = 0x886F
	CDMA_2000               Type = 0x8881
	PROFINET                Type = 0x8892
	HYPERSCSI               Type = 0x889a
	AOE                     Type = 0x88a2
	BRDWALK                 Type = 0x88AE
	LOOPBACK                Type = 0x9000
	RTNET_MAC               Type = 0x9021
	RTNET_CONFIG            Type = 0x9022
	PGLAN                   Type = 0x9999
	SRP_ISIS                Type = 0xFEFE
	RESERVED                Type = 0xFFFF
)

var typeStrings = [...]string{
	XNS_IDP:                 "XNS_IDP",
	IP4:                     "IP4",
	ARP:                     "ARP",
	VINES_IP:                "VINES_IP",
	VINES_LOOPBACK:          "VINES_LOOPBACK",
	VINES_ECHO:              "VINES_ECHO",
	TRAIN:                   "TRAIN",
	CDP:                     "CDP",
	CGMP:                    "CGMP",
	SRP_CONTROL:             "SRP_CONTROL",
	CENTRINO_PROMISC:        "CENTRINO_PROMISC",
	DECNET:                  "DECNET",
	DECNET_DUMP_LOAD:        "DECNET_DUMP_LOAD",
	DECNET_REMOTE_CONSOLE:   "DECNET_REMOTE_CONSOLE",
	DECNET_ROUTE:            "DECNET_ROUTE",
	DEC_LAT:                 "DEC_LAT",
	DEC_DIAGNOSTIC:          "DEC_DIAGNOSTIC",
	DEC_CUSTOMER:            "DEC_CUSTOMER",
	DEC_SCA:                 "DEC_SCA",
	TRANSPARENT_BRIDGING:    "TRANSPARENT_BRIDGING",
	RAW_FRAME_RELAY:         "RAW_FRAME_RELAY",
	REVERSE_ARP:             "REVERSE_ARP",
	DEC_LAN_BRIDGE:          "DEC_LAN_BRIDGE",
	DEC_ETHERNET_ENCRYPTION: "DEC_ETHERNET_ENCRYPTION",
	DEC_LAN_TRAFFIC_MONITOR: "DEC_LAN_TRAFFIC_MONITOR",
	DEC_LAST:                "DEC_LAST",
	APPLETALK:               "APPLETALK",
	IBM_SNA:                 "IBM_SNA",
	APPLETALK_AARP:          "APPLETALK_AARP",
	WELLFLEET_COMPRESSION:   "WELLFLEET_COMPRESSION",
	VLAN:                   "VLAN",
	VLAN_IN_VLAN:           "VLAN_IN_VLAN",
	IPX:                    "IPX",
	SNMP:                   "SNMP",
	CABLETRON_ISMP:         "CABLETRON_ISMP",
	CABLETRON_ISMP_TBFLOOD: "CABLETRON_ISMP_TBFLOOD",
	IP6:                     "IP6",
	ATOMIC:                  "ATOMIC",
	TCP_IP_COMPRESSION:      "TCP_IP_COMPRESSION",
	IP_AUTONOMOUS_SYSTEMS:   "IP_AUTONOMOUS_SYSTEMS",
	SECURE_DATA:             "SECURE_DATA",
	MAC_CONTROL:             "MAC_CONTROL",
	SLOW_PROTOCOLS:          "SLOW_PROTOCOLS",
	PPP:                     "PPP",
	MPLS_UNICAST:            "MPLS_UNICAST",
	MPLS_MULTICAST:          "MPLS_MULTICAST",
	PPPOE_DISCOVERY:         "PPPOE_DISCOVERY",
	PPPOE_SESSION:           "PPPOE_SESSION",
	INTEL_ANS:               "INTEL_ANS",
	MICROSOFT_NLB_HEARTBEAT: "MICROSOFT_NLB_HEARTBEAT",
	CDMA_2000:               "CDMA_2000",
	PROFINET:                "PROFINET",
	HYPERSCSI:               "HYPERSCSI",
	AOE:                     "AOE",
	BRDWALK:                 "BRDWALK",
	LOOPBACK:                "LOOPBACK",
	RTNET_MAC:               "RTNET_MAC",
	RTNET_CONFIG:            "RTNET_CONFIG",
	PGLAN:                   "PGLAN",
	SRP_ISIS:                "SRP_ISIS",
	RESERVED:                "RESERVED",
}

func (t Type) String() string {
	return elib.StringerHex(typeStrings[:], int(t))
}

var typeMap = parse.NewStringMap(typeStrings[:])

func (t *Type) Parse(in *parse.Input) {
	var v uint16
	if !in.Parse("%v", typeMap, &v) {
		panic(parse.ErrInput)
	}
	*t = Type(v).FromHost()
}
