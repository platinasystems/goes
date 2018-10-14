// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bgp

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

type bgp struct {
	docker.Docket
}

var Suite = test.Suite{
	Name: "bgp",
	Tests: test.Tests{
		&bgp{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/bird/bgp/conf.yaml.tmpl",
			},
		},
		&bgp{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/bird/bgp/vlan/conf.yaml.tmpl",
			},
		},
	},
}

func (bgp *bgp) Test(t *testing.T) {
	bgp.Docket.Tests = test.Tests{
		&test.Unit{"check connectivity", bgp.checkConnectivity},
		&test.Unit{"check bird", bgp.checkBird},
		&test.Unit{"check neighbors", bgp.checkNeighbors},
		&test.Unit{"check routes", bgp.checkRoutes},
		&test.Unit{"check interconnectivity",
			bgp.checkInterConnectivity},
		&test.Unit{"check flap", bgp.checkFlap},
		&test.Unit{"check connectivity2", bgp.checkConnectivity},
		&test.Unit{"check admin down", bgp.adminDown},
	}
	bgp.Docket.Test(t)
}

func (bgp *bgp) checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		host   string
		target string
	}{
		{"R1", "192.168.120.10"},
		{"R1", "192.168.150.4"},
		{"R2", "192.168.222.2"},
		{"R2", "192.168.120.5"},
		{"R3", "192.168.222.10"},
		{"R3", "192.168.111.4"},
		{"R4", "192.168.111.2"},
		{"R4", "192.168.150.5"},
	} {
		err := bgp.PingCmd(t, x.host, x.target)
		assert.Nil(err)
	}
}

func (bgp *bgp) checkBird(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	for _, r := range bgp.Routers {
		assert.Comment("Checking BIRD on", r.Hostname)
		out, err := bgp.ExecCmd(t, r.Hostname, "ps", "ax")
		assert.Nil(err)
		assert.Match(out, ".*bird.*")
	}
}

func (bgp *bgp) checkNeighbors(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		peer     string
	}{
		{"R1", "R2"},
		{"R1", "R4"},
		{"R2", "R1"},
		{"R2", "R3"},
		{"R3", "R2"},
		{"R3", "R4"},
		{"R4", "R1"},
		{"R4", "R3"},
	} {
		found := false
		timeout := 120

		for i := timeout; i > 0; i-- {
			out, err := bgp.ExecCmd(t, x.hostname,
				"birdc", "show", "protocols", "all", x.peer)
			assert.Nil(err)
			if !assert.MatchNonFatal(out, ".*Established.*") {
				time.Sleep(1 * time.Second)
			} else {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("No bgp peer established for %v", x.hostname)
		}
	}
}

func (bgp *bgp) checkRoutes(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		route    string
	}{
		{"R1", "192.168.222.0/24"},
		{"R1", "192.168.111.0/24"},
		{"R2", "192.168.150.0/24"},
		{"R2", "192.168.111.0/24"},
		{"R3", "192.168.120.0/24"},
		{"R3", "192.168.150.0/24"},
		{"R4", "192.168.120.0/24"},
		{"R4", "192.168.222.0/24"},
	} {
		found := false
		timeout := 60
		for i := timeout; i > 0; i-- {
			out, err := bgp.ExecCmd(t, x.hostname,
				"ip", "route", "show", x.route)
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

func (bgp *bgp) checkInterConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		target   string
	}{
		{"R1", "192.168.222.2"},
		{"R1", "192.168.111.2"},
		{"R2", "192.168.111.4"},
		{"R2", "192.168.150.4"},
		{"R3", "192.168.120.5"},
		{"R3", "192.168.150.5"},
		{"R4", "192.168.120.10"},
		{"R4", "192.168.222.10"},
	} {
		err := bgp.PingCmd(t, x.hostname, x.target)
		assert.Nil(err)
		assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
	}
}

func (bgp *bgp) checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range bgp.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := bgp.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			_, err = bgp.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "up", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}

func (bgp *bgp) adminDown(t *testing.T) {
	assert := test.Assert{t}

	num_intf := 0
	for _, r := range bgp.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := bgp.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			num_intf++
		}
	}
	err := test.NoAdjacency(t)
	assert.Nil(err)
}
