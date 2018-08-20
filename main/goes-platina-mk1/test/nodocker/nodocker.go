// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nodocker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

type route struct {
	prefix string
	gw     string
}

type config struct {
	netns   string
	ifname  string
	ifa     string
	routes  []route
	remotes []string
}

type nodocker map[string]*config

var Suite = nodocker{
	"net0port0": &config{
		netns: "h1",
		ifa:   "10.1.0.0/31",
		routes: []route{
			{"10.1.0.2/31", "10.1.0.1"},
		},
		remotes: []string{"10.1.0.2"},
	},
	"net0port1": &config{
		netns: "r",
		ifa:   "10.1.0.1/31",
	},
	"net1port0": &config{
		netns: "h2",
		ifa:   "10.1.0.2/31",
		routes: []route{
			{"10.1.0.0/31", "10.1.0.3"},
		},
		remotes: []string{"10.1.0.0"},
	},
	"net1port1": &config{
		netns: "r",
		ifa:   "10.1.0.3/31",
	},
}

func (nodocker) String() string { return "nodocker" }

func (m nodocker) Test(t *testing.T) {
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}
	for k, c := range m {
		ifname := netport.PortByNetPort[k]
		c.ifname = ifname
		ns := c.netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err != nil {
			assert.Program(test.Self{},
				"ip", "netns", "add", ns)
			defer cleanup.Program(test.Self{},
				"ip", "netns", "del", ns)
		}
		assert.Program(test.Self{},
			"ip", "link", "set", ifname, "up", "netns", ns)
		defer cleanup.Program(test.Self{},
			"ip", "netns", "exec", ns, test.Self{},
			"ip", "link", "set", ifname, "down", "netns", 1)
		assert.Program(test.Self{},
			"ip", "netns", "exec", ns, test.Self{},
			"ip", "address", "add", c.ifa, "dev", ifname)
		defer cleanup.Program(test.Self{},
			"ip", "netns", "exec", ns, test.Self{},
			"ip", "address", "del", c.ifa, "dev", ifname)
		for _, route := range c.routes {
			prefix := route.prefix
			gw := route.gw
			assert.Program(test.Self{},
				"ip", "netns", "exec", ns, test.Self{},
				"ip", "route", "add", prefix, "via", gw)
		}
	}
	for _, c := range m {
		assert.Nil(test.Carrier(c.netns, c.ifname))
	}
	test.Tests{
		&test.Unit{"ping-gateways", m.pingGateways},
		&test.Unit{"ping-remotes", m.pingRemotes},
		&test.Unit{"hping01", m.hping01},
		&test.Unit{"hping10", m.hping10},
		&test.Unit{"hping30", m.hping30},
		&test.Unit{"hping60", m.hping60},
	}.Test(t)
}

func (m nodocker) pingGateways(t *testing.T) {
	assert := test.Assert{t}
	for _, c := range m {
		for _, r := range c.routes {
			assert.Nil(test.Ping(c.netns, r.gw))
		}
	}
}

func (m nodocker) pingRemotes(t *testing.T) {
	assert := test.Assert{t}
	for _, c := range m {
		for _, r := range c.remotes {
			assert.Nil(test.Ping(c.netns, r))
		}
	}
}

func (m nodocker) hping01(t *testing.T) { m.hping(t, 1*time.Second) }
func (m nodocker) hping10(t *testing.T) { m.hping(t, 10*time.Second) }
func (m nodocker) hping30(t *testing.T) { m.hping(t, 30*time.Second) }
func (m nodocker) hping60(t *testing.T) { m.hping(t, 60*time.Second) }

func (m nodocker) hping(t *testing.T, duration time.Duration) {
	assert := test.Assert{t}
	c := m["net0port0"]
	ns := c.netns
	gw := c.routes[0].gw
	assert.Nil(test.Ping(ns, gw))
	p, err := test.Begin(t, duration, test.Quiet{}, test.Self{},
		"ip", "netns", "exec", ns,
		"hping3", "--icmp", "--flood", "-q", "-t", 1, gw)
	assert.Nil(err)
	p.End()
	assert.Nil(test.Ping(ns, gw))
}
