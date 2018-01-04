// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/ip/route/mod"
	"github.com/platinasystems/go/goes/cmd/ip/route/show"
	"github.com/platinasystems/go/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "route",
	USAGE: `
	ip route [ show ]
	ip route { show | flush } SELECTOR
	ip route save SELECTOR
	ip route restore
	ip route { add | del | change | append | replace } ROUTE
	ip route get ADDRESS [ from ADDRESS iif IFNAME  ] [ oif IFNAME ]
		[ tos TOS ] [ vrf NAME ]

SELECTOR := [ root PREFIX ] [ match PREFIX ] [ exact PREFIX ]
	[ table TABLE_ID ] [ vrf NAME ] [ proto RTPROTO ]
	[ type TYPE ] [ scope SCOPE ]

ROUTE := NODE_SPEC [ INFO_SPEC ]

NODE_SPEC := [ TYPE ] PREFIX [ tos TOS ] [ table TABLE_ID ]
	[ proto RTPROTO ] [ scope SCOPE ] [ metric METRIC ]

INFO_SPEC := NH OPTIONS [ nexthop NH ] ...

NH := [ encap ENCAP ] [ via ADDRESS ] [ dev STRING ] [ weight NUMBER ]
	NHFLAGS

OPTIONS := [ mtu NUMBER ] [ advmss NUMBER ] [ as [ to ] ADDRESS ]
	[ rtt TIME ] [ rttvar TIME ] [ reordering NUMBER ]
	[ window NUMBER ] [ cwnd NUMBER ] [ ssthresh REALM ]
	[ realms REALM ] [ rto_min TIME ] [ initcwnd NUMBER ]
	[ initrwnd NUMBER ] [ features FEATURES ] [ quickack BOOL ]
	[ congctl NAME ] [ pref PREF ] [ expires TIME ]

TYPE := [ unicast | local | broadcast | multicast | throw | unreachable
       | prohibit | blackhole | nat ]

TABLE_ID := [ local| main | default | all | NUMBER ]

SCOPE := [ host | link | global | NUMBER ]

NHFLAGS := [ onlink | pervasive ]

RTPROTO := [ kernel | boot | static | NUMBER ]

FEATURES := [ ecn | ]

PREF := [ low | medium | high ]

ENCAP := [ MPLS | IP ]

ENCAP_MPLS := mpls [ LABEL ]

ENCAP_IP := ip id TUNNEL_ID dst REMOTE_IP [ tos TOS ] [ ttl TTL ]`,
	APROPOS: lang.Alt{
		lang.EnUS: "routing table management",
	},
	MAN: lang.Alt{
		lang.EnUS: Man,
	},
	ByName: map[string]cmd.Cmd{
		"add":     mod.Command("add"),
		"append":  mod.Command("append"),
		"change":  mod.Command("change"),
		"delete":  mod.Command("delete"),
		"replace": mod.Command("replace"),
		"":        show.Command(""),
		"show":    show.Command("show"),
		"flush":   show.Command("flush"),
		"get":     show.Command("get"),
		"save":    show.Command("save"),
		"restore": show.Command("restore"),
	},
}
