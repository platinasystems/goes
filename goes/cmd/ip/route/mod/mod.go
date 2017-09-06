// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "route table entry"
	Man     = `
SEE ALSO
	ip man route || ip route -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		// TYPE
		"unicast",
		"local",
		"broadcast",
		"multicast",
		"throw",
		"unreachable",
		"prohibit",
		"blackhole",
		"nat",
		// NHFLAGS
		"onlink",
		"pervasive",
	}
	Parms = []interface{}{
		// NODE_SPEC
		"tos",
		"table",
		"proto",
		"scope",
		"metric",
		// NH
		"encap",
		"via",
		"dev",
		"weight",
		// OPTIONS
		"mtu",
		"advmss",
		"as", // FIXME what about [ to ] ?
		"rtt",
		"rttvar",
		"reordering",
		"window",
		"cwnd",
		"ssthresh",
		"realms",
		"rto_min",
		"initcwnd",
		"initrwnd",
		"features",
		"quickac",
		"congctl",
		"pref",
		"expires",
	}
)

func New(s string) Command { return Command(s) }

type Command string

type mod options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (c Command) Usage() string {
	return fmt.Sprint("ip route ", c, ` NODE-SPEC [ INFO-SPEC ]

NODE-SPEC := [ TYPE ] PREFIX [ tos TOS ] [ table TABLE-ID ]
	[ proto RTPROTO ] [ scope SCOPE ] [ metric METRIC ]

INFO-SPEC := NH OPTIONS FLAGS [ nexthop NH ] ...

NH := [ encap ENCAP ] [ via [ FAMILY ] ADDRESS ] [ dev STRING ]
	[ weight NUMBER ] [ NHFLAGS ]

FAMILY := { inet | inet6 | ipx | dnet | mpls | bridge | link }

OPTIONS := FLAGS [ mtu NUMBER ] [ advmss NUMBER ] [ as [ to ] ADDRESS ]
	[ rtt TIME ] [ rttvar TIME ] [ reordering NUMBER ]
	[ window NUMBER ] [ cwnd NUMBER ] [ ssthresh REALM ]
	[ realms REALM ] [ rto_min TIME ] [ initcwnd NUMBER ]
	[ initrwnd NUMBER ] [ features FEATURES ] [ quickack BOOL ]
	[ congctl NAME ] [ pref PREF ] [ expires TIME ]

TYPE := { unicast | local | broadcast | multicast | throw | unreachable |
	prohibit | blackhole | nat }

TABLE-ID := { local| main | default | all | NUMBER }

SCOPE := { host | link | global | NUMBER }

NHFLAGS := { onlink | pervasive }

RTPROTO := { kernel | boot | static | NUMBER }

FEATURES := { ecn }

PREF := { low | medium | high }

ENCAP := { ENCAP-MPLS | ENCAP-IP }

ENCAP-MPLS := mpls [ LABEL ]

ENCAP-IP := ip id TUNNEL-ID dst REMOTE-IP [ tos TOS ] [ ttl TTL ]`)
}

func (c Command) Main(args ...string) error {
	var (
		err   error
		ip    net.IP
		ipnet *net.IPNet
	)

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	mod := (*mod)(o)
	args = mod.Flags.More(args, Flags...)
	args = mod.Parms.More(args, Parms...)

	if len(args) == 0 {
		return fmt.Errorf("PREFIX: missing")
	}

	ip, ipnet, err = net.ParseCIDR(args[0])
	if err != nil {
		return err
	}
	args = args[1:]

	// FIXME what about "nexthop NH"...?

	fmt.Println("FIXME", c, ip, ipnet)

	return nil
}
