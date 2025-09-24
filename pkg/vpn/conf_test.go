// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/kvc"
)

func TestAdmins(t *testing.T) {
	input := strings.NewReader(`
admin1	# tab comment
admin2  # space comment
`)
	reg := newRegistry()
	err := kvc.Range(input, reg.loadAdminsKeyValues)
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
	input := strings.NewReader(`
fc00:1234::a hosta
fc00:1234::b hostb
fc00:1234::c hostc
`)
	Prefix.Value = netip.MustParsePrefix("fc00:1234::/64")
	HostsFile.Value = "TestHost"
	reg := newRegistry()
	err := kvc.Range(input, reg.loadHostsKeyValues)
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
	input := strings.NewReader(`
1.2.3.4.5 hosta
`)
	Prefix.Value = netip.MustParsePrefix("1.2.3.4/32")
	HostsFile.Value = "TestHostBadIP4"
	reg := newRegistry()
	err := kvc.Range(input, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}

func TestHostsBadIP6(t *testing.T) {
	input := strings.NewReader(`
fc00:1234::x hosta
`)
	Prefix.Value = netip.MustParsePrefix("fc00:1234::/64")
	HostsFile.Value = "TestHostBadIP6"
	reg := newRegistry()
	err := kvc.Range(input, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}

func TestHostsRangeIP6(t *testing.T) {
	input := strings.NewReader(`
fc00:5678::a hosta
`)
	Prefix.Value = netip.MustParsePrefix("fc00:1234::/64")
	HostsFile.Value = "TestHostRangeIP6"
	reg := newRegistry()
	err := kvc.Range(input, reg.loadHostsKeyValues)
	if err == nil {
		t.Fatal("uncaught bad address")
	} else {
		t.Log("caught:", err)
	}
}
