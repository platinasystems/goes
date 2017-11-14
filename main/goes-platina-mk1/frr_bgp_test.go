// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
)

var bgp_config *Config

func FrrBGP(t *testing.T, confFile string) {

	bgp_config = LaunchContainers(t, confFile)
	defer TearDownContainers(t, bgp_config)

	Suite{
		{"l2", bgpCheckL2Connectivity},
		{"frr", checkBgpRunning},
		{"neighbors", checkBgpNeighbors},
		{"routes", checkBgpLearnedRoute},
		{"ping-learned", checkBgpConnectivityLearned},
	}.Run(t)
}

func bgpCheckL2Connectivity(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.120.10",
	).Output("/1 received/").Done()

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.150.4",
	).Output("/1 received/").Done()

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R2",
		"ping", "-c1", "192.168.222.2",
	).Output("/1 received/").Done()

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R3",
		"ping", "-c1", "192.168.111.4",
	).Output("/1 received/").Done()

	Assert{t}.Program(nil,
		"goes", "vnet", "show", "ip", "fib",
	).Ok().Done()
}

func checkBgpRunning(t *testing.T) {
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range bgp_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := DockerExecCmd(t, r.Hostname, bgp_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/bgpd/").Done()
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/zebra/").Done()
	}
}

func checkBgpNeighbors(t *testing.T) {
	time.Sleep(60 * time.Second) // give bgp time to converge

	cmd := []string{"vtysh", "-c", "show ip bgp neighbor 192.168.120.10"}
	t.Log(cmd)
	out, err := DockerExecCmd(t, "R1", bgp_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/state = Established/").Done()

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.222.2"}
	t.Log(cmd)
	out, err = DockerExecCmd(t, "R2", bgp_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/state = Established/").Done()

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.111.4"}
	t.Log(cmd)
	out, err = DockerExecCmd(t, "R3", bgp_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/state = Established/").Done()

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.150.5"}
	t.Log(cmd)
	out, err = DockerExecCmd(t, "R4", bgp_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/state = Established/").Done()
}

func checkBgpLearnedRoute(t *testing.T) {
	cmd := []string{"ip", "route", "show"}
	out, err := DockerExecCmd(t, "R1", bgp_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/192.168.222.0/24/").Done()
}

func checkBgpConnectivityLearned(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2",
	).Output("/1 received/").Done()
	Assert{t}.Program(nil,
		"goes", "vnet", "show", "ip", "fib",
	).Ok().Done()
}
