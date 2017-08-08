// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

// 8-bit protocol field from IP 4/6 headers.
type Protocol uint8

const (
	IP6_HOP_BY_HOP_OPTIONS Protocol = 0
	ICMP                   Protocol = 1
	IGMP                   Protocol = 2
	GGP                    Protocol = 3
	IP_IN_IP               Protocol = 4
	ST                     Protocol = 5
	TCP                    Protocol = 6
	CBT                    Protocol = 7
	EGP                    Protocol = 8
	IGP                    Protocol = 9
	BBN_RCC_MON            Protocol = 10
	NVP2                   Protocol = 11
	PUP                    Protocol = 12
	ARGUS                  Protocol = 13
	EMCON                  Protocol = 14
	XNET                   Protocol = 15
	CHAOS                  Protocol = 16
	UDP                    Protocol = 17
	MUX                    Protocol = 18
	DCN_MEAS               Protocol = 19
	HMP                    Protocol = 20
	PRM                    Protocol = 21
	XNS_IDP                Protocol = 22
	TRUNK1                 Protocol = 23
	TRUNK2                 Protocol = 24
	LEAF1                  Protocol = 25
	LEAF2                  Protocol = 26
	RDP                    Protocol = 27
	IRTP                   Protocol = 28
	ISO_TP4                Protocol = 29
	NETBLT                 Protocol = 30
	MFE_NSP                Protocol = 31
	MERIT_INP              Protocol = 32
	SEP                    Protocol = 33
	THREE_PC               Protocol = 34
	IDPR                   Protocol = 35
	XTP                    Protocol = 36
	DDP                    Protocol = 37
	IDPR_CMTP              Protocol = 38
	TP                     Protocol = 39
	IL                     Protocol = 40
	IP6_IN_IP              Protocol = 41
	SDRP                   Protocol = 42
	IP6_ROUTE              Protocol = 43
	IP6_FRAG               Protocol = 44
	IDRP                   Protocol = 45
	RSVP                   Protocol = 46
	GRE                    Protocol = 47
	MHRP                   Protocol = 48
	BNA                    Protocol = 49
	IPSEC_ESP              Protocol = 50
	IPSEC_AH               Protocol = 51
	I_NLSP                 Protocol = 52
	SWIPE                  Protocol = 53
	NARP                   Protocol = 54
	MOBILE                 Protocol = 55
	TLSP                   Protocol = 56
	SKIP                   Protocol = 57
	ICMP6                  Protocol = 58
	IP6_NONXT              Protocol = 59
	IP6_DST_OPTIONS        Protocol = 60
	CFTP                   Protocol = 62
	SAT_EXPAK              Protocol = 64
	KRYPTOLAN              Protocol = 65
	RVD                    Protocol = 66
	IPPC                   Protocol = 67
	SAT_MON                Protocol = 69
	VISA                   Protocol = 70
	IPCV                   Protocol = 71
	CPNX                   Protocol = 72
	CPHB                   Protocol = 73
	WSN                    Protocol = 74
	PVP                    Protocol = 75
	BR_SAT_MON             Protocol = 76
	SUN_ND                 Protocol = 77
	WB_MON                 Protocol = 78
	WB_EXPAK               Protocol = 79
	ISO_IP                 Protocol = 80
	VMTP                   Protocol = 81
	SECURE_VMTP            Protocol = 82
	VINES                  Protocol = 83
	TTP                    Protocol = 84
	NSFNET_IGP             Protocol = 85
	DGP                    Protocol = 86
	TCF                    Protocol = 87
	EIGRP                  Protocol = 88
	OSPF                   Protocol = 89
	SPRITE_RPC             Protocol = 90
	LARP                   Protocol = 91
	MTP                    Protocol = 92
	AX                     Protocol = 93
	IPIP                   Protocol = 94
	MICP                   Protocol = 95
	SCC_SP                 Protocol = 96
	ETHERIP                Protocol = 97
	ENCAP                  Protocol = 98
	GMTP                   Protocol = 100
	IFMP                   Protocol = 101
	PNNI                   Protocol = 102
	PIM                    Protocol = 103
	ARIS                   Protocol = 104
	SCPS                   Protocol = 105
	QNX                    Protocol = 106
	A                      Protocol = 107
	IPCOMP                 Protocol = 108
	SNP                    Protocol = 109
	COMPAQ_PEER            Protocol = 110
	IPX_IN_IP              Protocol = 111
	VRRP                   Protocol = 112
	PGM                    Protocol = 113
	L2TP                   Protocol = 115
	DDX                    Protocol = 116
	IATP                   Protocol = 117
	STP                    Protocol = 118
	SRP                    Protocol = 119
	UTI                    Protocol = 120
	SMP                    Protocol = 121
	SM                     Protocol = 122
	PTP                    Protocol = 123
	ISIS                   Protocol = 124
	FIRE                   Protocol = 125
	CRTP                   Protocol = 126
	CRUDP                  Protocol = 127
	SSCOPMCE               Protocol = 128
	IPLT                   Protocol = 129
	SPS                    Protocol = 130
	PIPE                   Protocol = 131
	SCTP                   Protocol = 132
	FC                     Protocol = 133
	RSVP_E2E_IGNORE        Protocol = 134
	MOBILITY               Protocol = 135
	UDP_LITE               Protocol = 136
	MPLS_IN_IP             Protocol = 137
	RESERVED               Protocol = 255
)

var protocolStrings = [...]string{
	IP6_HOP_BY_HOP_OPTIONS: "IP6-HOP-BY-HOP-OPTIONS",
	ICMP:            "ICMP",
	IGMP:            "IGMP",
	GGP:             "GGP",
	IP_IN_IP:        "IP-IN-IP",
	ST:              "ST",
	TCP:             "TCP",
	CBT:             "CBT",
	EGP:             "EGP",
	IGP:             "IGP",
	BBN_RCC_MON:     "BBN-RCC-MON",
	NVP2:            "NVP2",
	PUP:             "PUP",
	ARGUS:           "ARGUS",
	EMCON:           "EMCON",
	XNET:            "XNET",
	CHAOS:           "CHAOS",
	UDP:             "UDP",
	MUX:             "MUX",
	DCN_MEAS:        "DCN-MEAS",
	HMP:             "HMP",
	PRM:             "PRM",
	XNS_IDP:         "XNS-IDP",
	TRUNK1:          "TRUNK1",
	TRUNK2:          "TRUNK2",
	LEAF1:           "LEAF1",
	LEAF2:           "LEAF2",
	RDP:             "RDP",
	IRTP:            "IRTP",
	ISO_TP4:         "ISO-TP4",
	NETBLT:          "NETBLT",
	MFE_NSP:         "MFE-NSP",
	MERIT_INP:       "MERIT-INP",
	SEP:             "SEP",
	THREE_PC:        "THREE-PC",
	IDPR:            "IDPR",
	XTP:             "XTP",
	DDP:             "DDP",
	IDPR_CMTP:       "IDPR-CMTP",
	TP:              "TP",
	IL:              "IL",
	IP6_IN_IP:       "IP6-IN-IP",
	SDRP:            "SDRP",
	IP6_ROUTE:       "IP6-ROUTE",
	IP6_FRAG:        "IP6-FRAG",
	IDRP:            "IDRP",
	RSVP:            "RSVP",
	GRE:             "GRE",
	MHRP:            "MHRP",
	BNA:             "BNA",
	IPSEC_ESP:       "IPSECESP",
	IPSEC_AH:        "IPSEC-AH",
	I_NLSP:          "I-NLSP",
	SWIPE:           "SWIPE",
	NARP:            "NARP",
	MOBILE:          "MOBILE",
	TLSP:            "TLSP",
	SKIP:            "SKIP",
	ICMP6:           "ICMP6",
	IP6_NONXT:       "IP6-NONXT",
	IP6_DST_OPTIONS: "IP6-DST-OPTIONS",
	CFTP:            "CFTP",
	SAT_EXPAK:       "SAT-EXPAK",
	KRYPTOLAN:       "KRYPTOLAN",
	RVD:             "RVD",
	IPPC:            "IPPC",
	SAT_MON:         "SAT-MON",
	VISA:            "VISA",
	IPCV:            "IPCV",
	CPNX:            "CPNX",
	CPHB:            "CPHB",
	WSN:             "WSN",
	PVP:             "PVP",
	BR_SAT_MON:      "BR-SAT-MON",
	SUN_ND:          "SUN-ND",
	WB_MON:          "WB-MON",
	WB_EXPAK:        "WB-EXPAK",
	ISO_IP:          "ISO-IP",
	VMTP:            "VMTP",
	SECURE_VMTP:     "SECURE-VMTP",
	VINES:           "VINES",
	TTP:             "TTP",
	NSFNET_IGP:      "NSFNET-IGP",
	DGP:             "DGP",
	TCF:             "TCF",
	EIGRP:           "EIGRP",
	OSPF:            "OSPF",
	SPRITE_RPC:      "SPRITE-RPC",
	LARP:            "LARP",
	MTP:             "MTP",
	AX:              "AX",
	IPIP:            "IPIP",
	MICP:            "MICP",
	SCC_SP:          "SCC-SP",
	ETHERIP:         "ETHERIP",
	ENCAP:           "ENCAP",
	GMTP:            "GMTP",
	IFMP:            "IFMP",
	PNNI:            "PNNI",
	PIM:             "PIM",
	ARIS:            "ARIS",
	SCPS:            "SCPS",
	QNX:             "QNX",
	A:               "A",
	IPCOMP:          "IPCOMP",
	SNP:             "SNP",
	COMPAQ_PEER:     "COMPAQ-PEER",
	IPX_IN_IP:       "IPX-IN-IP",
	VRRP:            "VRRP",
	PGM:             "PGM",
	L2TP:            "L2TP",
	DDX:             "DDX",
	IATP:            "IATP",
	STP:             "STP",
	SRP:             "SRP",
	UTI:             "UTI",
	SMP:             "SMP",
	SM:              "SM",
	PTP:             "PTP",
	ISIS:            "ISIS",
	FIRE:            "FIRE",
	CRTP:            "CRTP",
	CRUDP:           "CRUDP",
	SSCOPMCE:        "SSCOPMCE",
	IPLT:            "IPLT",
	SPS:             "SPS",
	PIPE:            "PIPE",
	SCTP:            "SCTP",
	FC:              "FC",
	RSVP_E2E_IGNORE: "RSVP-E2E-IGNORE",
	MOBILITY:        "MOBILITY",
	UDP_LITE:        "UDP-LITE",
	MPLS_IN_IP:      "MPLS-IN-IP",
	RESERVED:        "RESERVED",
}

func (p Protocol) String() string {
	return elib.StringerHex(protocolStrings[:], int(p))
}

func (v Protocol) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(Protocol)
	if m == 0xff {
		return v.String()
	}
	return fmt.Sprintf("0x%x/%x", v, m)
}

var protocolMap = parse.NewStringMap(protocolStrings[:])

func (p *Protocol) Parse(in *parse.Input) { in.Parse("%v", protocolMap, p) }
