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

var config *docker.Config

func Test(t *testing.T, yaml []byte) {
	var err error
	config, err = docker.LaunchContainers(t, yaml)
	if err != nil {
		t.Fatalf("Error launchContainers: %v", err)
	}
	defer docker.TearDownContainers(t, config)

	test.Suite{
		{"connectivity", checkConnectivity},
		{"frr", checkFrr},
		{"neighbors", checkNeighbors},
		{"routes", checkRoutes},
		{"inter-connectivity", checkInterConnectivity},
		{"flap", checkFlap},
	}.Run(t)
}

func checkConnectivity(t *testing.T) {
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
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.host, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
	}
}

func checkFrr(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, ".*bgpd.*")
		assert.Match(out, ".*zebra.*")
	}
}

func checkNeighbors(t *testing.T) {
	assert := test.Assert{t}

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
		timeout := 120

		cmd := "show ip bgp neighbor " + x.peer
		vcmd := []string{"vtysh", "-c", cmd}
		for i := timeout; i > 0; i-- {
			out, err := docker.ExecCmd(t, x.hostname, config, vcmd)
			assert.Nil(err)
			if !assert.MatchNonFatal(out, ".*state = Established.*") {
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

func checkRoutes(t *testing.T) {
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
		cmd := []string{"ip", "route", "show", x.route}
		timeout := 60
		for i := timeout; i > 0; i-- {
			out, err := docker.ExecCmd(t, x.hostname, config, cmd)
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

func checkInterConnectivity(t *testing.T) {
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
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
	}
}

func checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range config.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			cmd := []string{"ip", "link", "set", "down", intf}
			_, err := docker.ExecCmd(t, r.Hostname, config, cmd)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			cmd = []string{"ip", "link", "set", "up", intf}
			_, err = docker.ExecCmd(t, r.Hostname, config, cmd)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}
