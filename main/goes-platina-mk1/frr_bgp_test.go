// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

var bgp_config *docker.Config

func FrrBGP(t *testing.T, confFile string) {

	bgp_config = docker.LaunchContainers(t, confFile)
	defer docker.TearDownContainers(t, bgp_config)

	test.Suite{
		{"l2", bgpCheckL2Connectivity},
		{"frr", checkBgpRunning},
		{"neighbors", checkBgpNeighbors},
		{"routes", checkBgpLearnedRoute},
		{"ping-learned", checkBgpConnectivityLearned},
	}.Run(t)
}

func bgpCheckL2Connectivity(t *testing.T) {
	assert := test.Assert{t}
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.120.10",
	)

	assert.Program(regexp.MustCompile("1 received.*"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.150.4",
	)

	assert.Program(regexp.MustCompile("1 received.*"),
		"goes", "ip", "netns", "exec", "R2",
		"ping", "-c1", "192.168.222.2",
	)

	assert.Program(regexp.MustCompile("1 received.*"),
		"goes", "ip", "netns", "exec", "R3",
		"ping", "-c1", "192.168.111.4",
	)

	assert.Program("goes", "vnet", "show", "ip", "fib")
}

func checkBgpRunning(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range bgp_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, bgp_config, cmd)
		assert.Nil(err)
		assert.Match(out, ".*bgpd.*")
		assert.Match(out, ".*zebra.*")
	}
}

func checkBgpNeighbors(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(60 * time.Second) // give bgp time to converge

	cmd := []string{"vtysh", "-c", "show ip bgp neighbor 192.168.120.10"}
	t.Log(cmd)
	out, err := docker.ExecCmd(t, "R1", bgp_config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*state = Established.*")

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.222.2"}
	t.Log(cmd)
	out, err = docker.ExecCmd(t, "R2", bgp_config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*state = Established.*")

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.111.4"}
	t.Log(cmd)
	out, err = docker.ExecCmd(t, "R3", bgp_config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*state = Established.*")

	cmd = []string{"vtysh", "-c", "show ip bgp neighbor 192.168.150.5"}
	t.Log(cmd)
	out, err = docker.ExecCmd(t, "R4", bgp_config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*state = Established.*")
}

func checkBgpLearnedRoute(t *testing.T) {
	assert := test.Assert{t}
	cmd := []string{"ip", "route", "show"}
	out, err := docker.ExecCmd(t, "R1", bgp_config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*192.168.222.0/24.*")
}

func checkBgpConnectivityLearned(t *testing.T) {
	assert := test.Assert{t}
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2",
	)
	assert.Program("goes", "vnet", "show", "ip", "fib")
}
