// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type bridge
//	[ fdb_flush ]
//	[ forward_delay FORWARD_DELAY ]
//	[ hello_time HELLO_TIME ]
//	[ max_age MAX_AGE ]
//	[ ageing_time AGEING_TIME ]
//	[ stp_state STP_STATE ]
//	[ priority PRIORITY ]
//	[ group_fwd_mask MASK ]
//	[ group_address ADDRESS ]
//	[ vlan_filtering VLAN_FILTERING ]
//	[ vlan_protocol VLAN_PROTOCOL ]
//	[ vlan_default_pvid VLAN_DEFAULT_PVID ]
//	[ vlan_stats_enabled VLAN_STATS_ENABLED ]
//	[ mcast_snooping MCAST_SNOOPING ]
//	[ mcast_router MCAST_ROUTER ]
//	[ mcast_query_use_ifaddr MCAST_QUERY_USE_IFADDR ]
//	[ mcast_querier MCAST_QUERIER ]
//	[ mcast_hash_elasticity HASH_ELASTICITY ]
//	[ mcast_hash_max HASH_MAX ]
//	[ mcast_last_member_count LAST_MEMBER_COUNT ]
//	[ mcast_startup_query_count STARTUP_QUERY_COUNT ]
//	[ mcast_last_member_interval LAST_MEMBER_INTERVAL ]
//	[ mcast_membership_interval MEMBERSHIP_INTERVAL ]
//	[ mcast_querier_interval QUERIER_INTERVAL ]
//	[ mcast_query_interval QUERY_INTERVAL ]
//	[ mcast_query_response_interval QUERY_RESPONSE_INTERVAL ]
//	[ mcast_startup_query_interval STARTUP_QUERY_INTERVAL ]
//	[ mcast_stats_enabled MCAST_STATS_ENABLED ]
//	[ mcast_igmp_version IGMP_VERSION ]
//	[ mcast_mld_version MLD_VERSION ]
//	[ nf_call_iptables NF_CALL_IPTABLES ]
//	[ nf_call_ip6tables NF_CALL_IP6TABLES ]
//	[ nf_call_arptables NF_CALL_ARPTABLES ]
func (m *mod) parseTypeBridge() error {
	var s string
	var err error

	m.args = m.opt.Flags.More(m.args, []string{"fdb-flush", "fdb_flush"})
	if m.opt.Flags.ByName["fdb-flush"] {
		m.attrs = append(m.attrs, rtnl.Attr{rtnl.IFLA_BR_FDB_FLUSH,
			rtnl.NilAttr{}})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"mcast-last-member-interval",
			"mcast_last_member_interval"},
			rtnl.IFLA_BR_MCAST_LAST_MEMBER_INTVL},
		{[]string{"mcast-membership-interval",
			"mcast_membership_interval"},
			rtnl.IFLA_BR_MCAST_MEMBERSHIP_INTVL},
		{[]string{"mcast-querier-interval",
			"mcast_querier_interval"},
			rtnl.IFLA_BR_MCAST_QUERIER_INTVL},
		{[]string{"mcast-query-interval", "mcast_query_interval"},
			rtnl.IFLA_BR_MCAST_QUERY_INTVL},
		{[]string{"mcast-query-response-interval",
			"mcast_query_response_interval"},
			rtnl.IFLA_BR_MCAST_QUERY_RESPONSE_INTVL},
		{[]string{"mcast-startup-query-interval",
			"mcast_startup_query_interval"},
			rtnl.IFLA_BR_MCAST_STARTUP_QUERY_INTVL},
	} {
		var u64 uint64
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u64); err != nil {
			return fmt.Errorf("%s: %q %v ", x.names[0], s, err)
		}
		m.attrs = append(m.attrs, rtnl.Attr{x.t,
			rtnl.Uint32Attr(u64)})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"forward-delay", "forward_delay"},
			rtnl.IFLA_BR_FORWARD_DELAY},
		{[]string{"hello-time", "hello_time"},
			rtnl.IFLA_BR_HELLO_TIME},
		{[]string{"max-age", "max-age"},
			rtnl.IFLA_BR_MAX_AGE},
		{[]string{"ageing-time", "ageing_time"},
			rtnl.IFLA_BR_AGEING_TIME},
		{[]string{"stp-state", "stp_state"},
			rtnl.IFLA_BR_STP_STATE},
		{[]string{"mcast-hash-elasticity", "mcast_hash_elasticity"},
			rtnl.IFLA_BR_MCAST_HASH_ELASTICITY},
		{[]string{"mcast-hash-max", "mcast_hash_max"},
			rtnl.IFLA_BR_MCAST_HASH_MAX},
		{[]string{"mcast-last-member-count",
			"mcast_last_member_count"},
			rtnl.IFLA_BR_MCAST_LAST_MEMBER_CNT},
		{[]string{"mcast-startup-query-count",
			"mcast_startup_query_count"},
			rtnl.IFLA_BR_MCAST_STARTUP_QUERY_CNT},
	} {
		var u32 uint32
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("%s: %q %v ", x.names[0], s, err)
		}
		m.attrs = append(m.attrs, rtnl.Attr{x.t,
			rtnl.Uint32Attr(u32)})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"priority"},
			rtnl.IFLA_BR_PRIORITY},
		{[]string{"vlan-protocol", "vlan_protocol"},
			rtnl.IFLA_BR_VLAN_PROTOCOL},
		{[]string{"group-fwd-mask", "group_fwd_mask"},
			rtnl.IFLA_BR_GROUP_FWD_MASK},
		{[]string{"vlan-default-pvid", "vlan_default_pvid"},
			rtnl.IFLA_BR_VLAN_DEFAULT_PVID},
	} {
		var u16 uint16
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("%s: %q %v ", x.names[0], s, err)
		}
		m.attrs = append(m.attrs, rtnl.Attr{x.t,
			rtnl.Uint16Attr(u16)})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"vlan-filtering", "vlan_filtering"},
			rtnl.IFLA_BR_VLAN_FILTERING},
		{[]string{"vlan-stats-enabled", "vlan_stats_enabled"},
			rtnl.IFLA_BR_VLAN_STATS_ENABLED},
		{[]string{"mcast-router", "mcast_router"},
			rtnl.IFLA_BR_MCAST_ROUTER},
		{[]string{"mcast-snooping", "mcast_snooping"},
			rtnl.IFLA_BR_MCAST_SNOOPING},
		{[]string{"mcast-query-use-ifaddr", "mcast_query_use_ifaddr"},
			rtnl.IFLA_BR_MCAST_QUERY_USE_IFADDR},
		{[]string{"mcast-querier", "mcast_querier"},
			rtnl.IFLA_BR_MCAST_QUERIER},
		{[]string{"mcast-stats-enabled", "mcast_stats_enabled"},
			rtnl.IFLA_BR_MCAST_STATS_ENABLED},
		{[]string{"mcast-igmp-version", "mcast_igmp_version"},
			rtnl.IFLA_BR_MCAST_IGMP_VERSION},
		{[]string{"mcast-mld-version", "mcast_mld_version"},
			rtnl.IFLA_BR_MCAST_MLD_VERSION},
		{[]string{"nf-call-iptables", "nf_call_iptables"},
			rtnl.IFLA_BR_NF_CALL_IPTABLES},
		{[]string{"nf-call-ip6tables", "nf_call_ip6tables"},
			rtnl.IFLA_BR_NF_CALL_IP6TABLES},
		{[]string{"nf-call-arptables", "nf_call_arptables"},
			rtnl.IFLA_BR_NF_CALL_ARPTABLES},
	} {
		var u8 uint8
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v ", x.names[0], s, err)
		}
		m.attrs = append(m.attrs, rtnl.Attr{x.t,
			rtnl.Uint16Attr(u8)})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"group-address", "group_address"},
			rtnl.IFLA_BR_GROUP_ADDR},
	} {
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		lladdr, err := net.ParseMAC(s)
		if err != nil {
			return fmt.Errorf("%s: %q %v", x.names[0], s, err)
		}
		m.attrs = append(m.attrs, rtnl.Attr{x.t,
			rtnl.BytesAttr(lladdr)})
	}
	return nil
}
