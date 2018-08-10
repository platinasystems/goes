// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ebgp

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

type ebgp struct {
	docker.Docket
}

var Suite = test.Suite{
	Name: "ebgp",
	Tests: test.Tests{
		&ebgp{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/gobgp/ebgp/conf.yaml.tmpl",
			},
		},
		&ebgp{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/gobgp/ebgp/vlan/conf.yaml.tmpl",
			},
		},
	},
}

func (ebgp *ebgp) Test(t *testing.T) {
	ebgp.Docket.Tests = test.Tests{
		&test.Unit{"check connectivity", ebgp.checkConnectivity},
		&test.Unit{"check gobgp", ebgp.checkBgp},
		&test.Unit{"check neighbor", ebgp.checkNeighbors},
		&test.Unit{"check route", ebgp.checkRoutes},
		&test.Unit{"check interconnectivity",
			ebgp.checkInterConnectivity},
		&test.Unit{"check flap", ebgp.checkFlap},
	}
	ebgp.Docket.Test(t)
}

func (ebgp *ebgp) checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		host   string
		target string
	}{
		{"R1", "192.168.120.10"},
		{"R1", "192.168.150.2"},
		{"R1", "192.168.1.5"},
		{"R2", "192.168.222.4"},
		{"R2", "192.168.120.5"},
		{"R2", "192.168.1.10"},
		{"R3", "192.168.150.5"},
		{"R3", "192.168.111.4"},
		{"R3", "192.168.2.2"},
		{"R4", "192.168.111.2"},
		{"R4", "192.168.222.10"},
		{"R4", "192.168.2.4"},
	} {
		out, err := ebgp.ExecCmd(t, x.host, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
	}
}

func (ebgp *ebgp) checkBgp(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	for _, r := range ebgp.Routers {
		t.Logf("Checking gobgp on %v", r.Hostname)
		out, err := ebgp.ExecCmd(t, r.Hostname, "ps", "ax")
		assert.Nil(err)
		assert.Match(out, ".*gobgpd.*")
		assert.Match(out, ".*zebra.*")
	}
}

func (ebgp *ebgp) checkNeighbors(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		peer     string
	}{
		{"R1", "192.168.120.10"},
		{"R1", "192.168.150.2"},
		{"R2", "192.168.120.5"},
		{"R2", "192.168.222.4"},
		{"R3", "192.168.150.5"},
		{"R3", "192.168.111.4"},
		{"R4", "192.168.222.10"},
		{"R4", "192.168.111.2"},
	} {
		found := false
		timeout := 120

		for i := timeout; i > 0; i-- {
			out, err := ebgp.ExecCmd(t, x.hostname,
				"/root/gobgp", "neighbor", x.peer)
			assert.Nil(err)
			if !assert.MatchNonFatal(out, ".*state = established.*") {
				time.Sleep(1 * time.Second)
			} else {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("No bgp peer established for %v", x.hostname)
		}
		_, err := ebgp.ExecCmd(t, x.hostname,
			"/root/gobgp", "global", "rib")
		assert.Nil(err)
	}
}

func (ebgp *ebgp) checkRoutes(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		route    string
	}{
		{"R1", "192.168.222.0/24"},
		{"R1", "192.168.111.0/24"},
		{"R1", "192.168.1.10/32"},
		{"R1", "192.168.2.2/32"},
		{"R1", "192.168.2.4/32"},

		{"R2", "192.168.150.0/24"},
		{"R2", "192.168.111.0/24"},
		{"R2", "192.168.1.5/32"},
		{"R2", "192.168.2.2/32"},
		{"R2", "192.168.2.4/32"},

		{"R3", "192.168.120.0/24"},
		{"R3", "192.168.222.0/24"},
		{"R3", "192.168.1.5/32"},
		{"R3", "192.168.1.10/32"},
		{"R3", "192.168.2.4/32"},

		{"R4", "192.168.120.0/24"},
		{"R4", "192.168.150.0/24"},
		{"R4", "192.168.1.5/32"},
		{"R4", "192.168.1.10/32"},
		{"R4", "192.168.2.2/32"},
	} {
		found := false
		timeout := 60
		for i := timeout; i > 0; i-- {
			out, err := ebgp.ExecCmd(t, x.hostname,
				"vtysh", "-c", "show ip route "+x.route)
			assert.Nil(err)
			if !assert.MatchNonFatal(out, x.route) {
				time.Sleep(1 * time.Second)
			} else {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("No bgp route for %v: %v", x.hostname, x.route)
		}
	}
}

func (ebgp *ebgp) checkInterConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		target   string
	}{

		{"R1", "192.168.222.4"},
		{"R1", "192.168.111.2"},
		{"R1", "192.168.111.4"},
		{"R1", "192.168.1.10"},
		{"R1", "192.168.2.2"},
		{"R1", "192.168.2.4"},

		{"R2", "192.168.111.4"},
		{"R2", "192.168.111.2"},
		{"R2", "192.168.150.2"},
		{"R2", "192.168.1.5"},
		{"R2", "192.168.2.2"},
		{"R2", "192.168.2.4"},

		{"R3", "192.168.120.5"},
		{"R3", "192.168.222.4"},
		{"R3", "192.168.1.5"},
		{"R3", "192.168.1.10"},
		{"R3", "192.168.2.4"},

		{"R4", "192.168.120.10"},
		{"R4", "192.168.150.2"},
		{"R4", "192.168.150.5"},
		{"R4", "192.168.1.5"},
		{"R4", "192.168.1.10"},
		{"R4", "192.168.2.2"},
	} {
		out, err := ebgp.ExecCmd(t, x.hostname, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
	}
}

func (ebgp *ebgp) checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range ebgp.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := ebgp.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			_, err = ebgp.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "up", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}
