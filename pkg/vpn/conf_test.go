// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net/netip"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/kvc"
)

func TestAdmins(t *testing.T) {
	const input = `
admin1	# tab comment
admin2  # space comment
`
	reg := newRegistry()
	err := kvc.RangeString(input, SplitConfLine, reg.loadAdminsKeyValues)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"admin1", "admin2"} {
		if _, ok := reg.admin[name]; !ok {
			t.Error(name, "not found")
		}
	}
}

func TestHosts(t *testing.T) {
	const input = `
fc00:1234::a hosta
fc00:1234::b hostb
fc00:1234::c hostc
`
	hostsFile = "test_hosts"
	prefix = netip.MustParsePrefix("fc00:1234::/64")
	reg := newRegistry()
	err := kvc.RangeString(input, SplitConfLine, reg.loadHostsKeyValues)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"hosta", "hostb", "hostc"} {
		if addr, ok := reg.hosts.addr[name]; !ok {
			t.Error(name, "not found")
		} else {
			t.Log(addr, name)
		}
	}
}

func TestHostsBadIP4(t *testing.T) {
	const input = "1.2.3.4.5 hosta"
	hostsFile = "test_hosts_invalid_ip4"
	prefix = netip.MustParsePrefix("1.2.3.4/32")
	reg := newRegistry()
	err := kvc.RangeString(input, SplitConfLine, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}

func TestHostsBadIP6(t *testing.T) {
	const input = "fc00:1234::x hosta"
	hostsFile = "test_hosts_invalid_ip6"
	prefix = netip.MustParsePrefix("fc00:1234::/64")
	reg := newRegistry()
	err := kvc.RangeString(input, SplitConfLine, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}

func TestHostsRangeIP6(t *testing.T) {
	const input = "fc00:5678::a hosta"
	hostsFile = "test_hosts_range_ip6"
	prefix = netip.MustParsePrefix("fc00:1234::/64")
	reg := newRegistry()
	err := kvc.RangeString(input, SplitConfLine, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}
