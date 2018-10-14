// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ospf

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

type ospf struct {
	docker.Docket
}

var Suite = test.Suite{
	Name: "ospf",
	Tests: test.Tests{
		&ospf{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/bird/ospf/conf.yaml.tmpl",
			},
		},
		&ospf{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/bird/ospf/vlan/conf.yaml.tmpl",
			},
		},
	},
}

func (ospf *ospf) Run(t *testing.T) {
	ospf.Docket.Tests = test.Tests{
		&test.Unit{"check connectivity", ospf.checkConnectivity},
		&test.Unit{"check bird", ospf.checkBird},
		&test.Unit{"check neighbors", ospf.checkNeighbors},
		&test.Unit{"check routes", ospf.checkRoutes},
		&test.Unit{"check interconnectivity",
			ospf.checkInterConnectivity},
		&test.Unit{"check flap", ospf.checkFlap},
		&test.Unit{"check connectivity2", ospf.checkConnectivity},
		&test.Unit{"check admin down", ospf.adminDown},
	}
	ospf.Docket.Test(t)
}

func (ospf *ospf) checkConnectivity(t *testing.T) {
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
		err := ospf.PingCmd(t, x.host, x.target)
		assert.Nil(err)
	}
}

func (ospf *ospf) checkBird(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range ospf.Routers {
		assert.Comment("Checking BIRD on", r.Hostname)
		out, err := ospf.ExecCmd(t, r.Hostname, "ps", "ax")
		assert.Nil(err)
		assert.Match(out, ".*bird.*")
	}
}

func (ospf *ospf) checkNeighbors(t *testing.T) {
	assert := test.Assert{t}

	timeout := 120

	for _, x := range []struct {
		hostname string
		peer     string
	}{
		{"R1", "192.168.120.10"},
		{"R1", "192.168.150.4"},
		{"R2", "192.168.120.5"},
		{"R2", "192.168.222.2"},
		{"R3", "192.168.222.10"},
		{"R3", "192.168.111.4"},
		{"R4", "192.168.111.2"},
		{"R4", "192.168.150.5"},
	} {
		found := false
		for i := timeout; i > 0; i-- {
			out, err := ospf.ExecCmd(t, x.hostname,
				"birdc", "show", "ospf", "neighbor")
			assert.Nil(err)
			if !assert.MatchNonFatal(out, x.peer) {
				time.Sleep(1 * time.Second)
			} else {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("No ospf neighbor found for %v", x.hostname)
		}
	}
}

func (ospf *ospf) checkRoutes(t *testing.T) {
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
			out, err := ospf.ExecCmd(t, x.hostname,
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
			t.Fatalf("No ospf route for %v: %v", x.hostname, x.route)
		}
	}
}

func (ospf *ospf) checkInterConnectivity(t *testing.T) {
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
		err := ospf.PingCmd(t, x.hostname, x.target)
		assert.Nil(err)
		assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
	}
}

func (ospf *ospf) checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range ospf.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := ospf.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			_, err = ospf.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "up", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}

func (ospf *ospf) adminDown(t *testing.T) {
	assert := test.Assert{t}

	num_intf := 0
	for _, r := range ospf.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := ospf.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			num_intf++
		}
	}
	err := test.NoAdjacency(t)
	assert.Nil(err)
}
