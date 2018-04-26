// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bridge

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "bridge" }

func (Command) Usage() string {
	return `
ip link add type bridge [ OPTIONS ]...`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a vlan virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	fdb-flush
	forward-delay FORWARD-DELAY
	hello-time HELLO-TIME
	max-age MAX-AGE
	ageing-time AGEING-TIME
	stp-state STP-STATE
	priority PRIORITY
	group-fwd-mask MASK
	group-address ADDRESS
	vlan-filtering VLAN-FILTERING
	vlan-protocol VLAN-PROTOCOL
	vlan-default-pvid VLAN-DEFAULT-PVID
	vlan-stats-enabled VLAN-STATS-ENABLED
	mcast-snooping MCAST-SNOOPING
	mcast-router MCAST-ROUTER
	mcast-query-use-ifaddr MCAST-QUERY-USE-IFADDR
	mcast-querier MCAST-QUERIER
	mcast-hash-elasticity HASH-ELASTICITY
	mcast-hash-max HASH-MAX
	mcast-last-member-count LAST-MEMBER-COUNT
	mcast-startup-query-count STARTUP-QUERY-COUNT
	mcast-last-member-interval LAST-MEMBER-INTERVAL
	mcast-membership-interval MEMBERSHIP-INTERVAL
	mcast-querier-interval QUERIER-INTERVAL
	mcast-query-interval QUERY-INTERVAL
	mcast-query-response-interval QUERY-RESPONSE-INTERVAL
	mcast-startup-query-interval STARTUP-QUERY-INTERVAL
	mcast-stats-enabled MCAST-STATS-ENABLED
	mcast-igmp-version IGMP-VERSION
	mcast-mld-version MLD-VERSION
	nf-call-iptables NF-CALL-IPTABLES
	nf-call-ip6tables NF-CALL-IP6TABLES
	nf-call-arptables NF-CALL-ARPTABLES

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var s string

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"fdb-flush", "fdb_flush"},
	)
	args = opt.Parms.More(args,
		[]string{"mcast-last-member-interval",
			"mcast_last_member_interval"},
		[]string{"mcast-membership-interval",
			"mcast_membership_interval"},
		[]string{"mcast-querier-interval",
			"mcast_querier_interval"},
		[]string{"mcast-query-interval",
			"mcast_query_interval"},
		[]string{"mcast-query-response-interval",
			"mcast_query_response_interval"},
		[]string{"mcast-startup-query-interval",
			"mcast_startup_query_interval"},
		[]string{"forward-delay", "forward_delay"},
		[]string{"hello-time", "hello_time"},
		[]string{"max-age", "max-age"},
		[]string{"ageing-time", "ageing_time"},
		[]string{"stp-state", "stp_state"},
		[]string{"mcast-hash-elasticity", "mcast_hash_elasticity"},
		[]string{"mcast-hash-max", "mcast_hash_max"},
		[]string{"mcast-last-member-count", "mcast_last_member_count"},
		[]string{"mcast-startup-query-count",
			"mcast_startup_query_count"},
		[]string{"priority"},
		[]string{"vlan-protocol", "vlan_protocol"},
		[]string{"group-fwd-mask", "group_fwd_mask"},
		[]string{"vlan-default-pvid", "vlan_default_pvid"},
		[]string{"vlan-filtering", "vlan_filtering"},
		[]string{"vlan-stats-enabled", "vlan_stats_enabled"},
		[]string{"mcast-router", "mcast_router"},
		[]string{"mcast-snooping", "mcast_snooping"},
		[]string{"mcast-querier", "mcast_querier"},
		[]string{"mcast-query-use-ifaddr", "mcast_query_use_ifaddr"},
		[]string{"mcast-stats-enabled", "mcast_stats_enabled"},
		[]string{"mcast-igmp-version", "mcast_igmp_version"},
		[]string{"mcast-mld-version", "mcast_mld_version"},
		[]string{"nf-call-iptables", "nf_call_iptables"},
		[]string{"nf-call-ip6tables", "nf_call_ip6tables"},
		[]string{"nf-call-arptables", "nf_call_arptables"},
		[]string{"group-address", "group_address"},
	)

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt, args)
	if err != nil {
		return err
	}

	if opt.Flags.ByName["fdb-flush"] {
		add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_BR_FDB_FLUSH,
			nl.NilAttr{}})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"mcast-last-member-interval",
			rtnl.IFLA_BR_MCAST_LAST_MEMBER_INTVL},
		{"mcast-membership-interval",
			rtnl.IFLA_BR_MCAST_MEMBERSHIP_INTVL},
		{"mcast-querier-interval",
			rtnl.IFLA_BR_MCAST_QUERIER_INTVL},
		{"mcast-query-interval",
			rtnl.IFLA_BR_MCAST_QUERY_INTVL},
		{"mcast-query-response-interval",
			rtnl.IFLA_BR_MCAST_QUERY_RESPONSE_INTVL},
		{"mcast-startup-query-interval",
			rtnl.IFLA_BR_MCAST_STARTUP_QUERY_INTVL},
	} {
		var u64 uint64
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u64); err != nil {
			return fmt.Errorf("%s: %q %v ", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.Uint32Attr(u64)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"forward-delay", rtnl.IFLA_BR_FORWARD_DELAY},
		{"hello-time", rtnl.IFLA_BR_HELLO_TIME},
		{"max-age", rtnl.IFLA_BR_MAX_AGE},
		{"ageing-time", rtnl.IFLA_BR_AGEING_TIME},
		{"stp-state", rtnl.IFLA_BR_STP_STATE},
		{"mcast-hash-elasticity", rtnl.IFLA_BR_MCAST_HASH_ELASTICITY},
		{"mcast-hash-max", rtnl.IFLA_BR_MCAST_HASH_MAX},
		{"mcast-last-member-count",
			rtnl.IFLA_BR_MCAST_LAST_MEMBER_CNT},
		{"mcast-startup-query-count",
			rtnl.IFLA_BR_MCAST_STARTUP_QUERY_CNT},
	} {
		var u32 uint32
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("%s: %q %v ", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.Uint32Attr(u32)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"priority", rtnl.IFLA_BR_PRIORITY},
		{"vlan-protocol", rtnl.IFLA_BR_VLAN_PROTOCOL},
		{"group-fwd-mask", rtnl.IFLA_BR_GROUP_FWD_MASK},
		{"vlan-default-pvid", rtnl.IFLA_BR_VLAN_DEFAULT_PVID},
	} {
		var u16 uint16
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("%s: %q %v ", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.Uint16Attr(u16)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"vlan-filtering", rtnl.IFLA_BR_VLAN_FILTERING},
		{"vlan-stats-enabled", rtnl.IFLA_BR_VLAN_STATS_ENABLED},
		{"mcast-router", rtnl.IFLA_BR_MCAST_ROUTER},
		{"mcast-snooping", rtnl.IFLA_BR_MCAST_SNOOPING},
		{"mcast-query-use-ifaddr", rtnl.IFLA_BR_MCAST_QUERY_USE_IFADDR},
		{"mcast-querier", rtnl.IFLA_BR_MCAST_QUERIER},
		{"mcast-stats-enabled", rtnl.IFLA_BR_MCAST_STATS_ENABLED},
		{"mcast-igmp-version", rtnl.IFLA_BR_MCAST_IGMP_VERSION},
		{"mcast-mld-version", rtnl.IFLA_BR_MCAST_MLD_VERSION},
		{"nf-call-iptables", rtnl.IFLA_BR_NF_CALL_IPTABLES},
		{"nf-call-ip6tables", rtnl.IFLA_BR_NF_CALL_IP6TABLES},
		{"nf-call-arptables", rtnl.IFLA_BR_NF_CALL_ARPTABLES},
	} {
		var u8 uint8
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v ", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.Uint16Attr(u8)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"group-address", rtnl.IFLA_BR_GROUP_ADDR},
	} {
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		lladdr, err := net.ParseMAC(s)
		if err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t,
			nl.BytesAttr(lladdr)})
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO,
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr("bridge")}})

	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
