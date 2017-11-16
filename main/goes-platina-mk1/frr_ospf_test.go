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

var ospf_config *docker.Config

func FrrOSPF(t *testing.T, confFile string) {

	ospf_config = docker.LaunchContainers(t, confFile)
	defer docker.TearDownContainers(t, ospf_config)

	test.Suite{
		{"l2", ospfCheckL2Connectivity},
		{"frr", checkOspfRunning},
		{"neighbors", checkOspfNeighbors},
		{"routes", checkOspfLearnedRoute},
		{"ping-learned", checkOspfConnectivityLearned},
	}.Run(t)
}

func ospfCheckL2Connectivity(t *testing.T) {
	assert := test.Assert{t}

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.120.10")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.150.4")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R2",
		"ping", "-c1", "192.168.222.2")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R3",
		"ping", "-c1", "192.168.111.4")

	assert.Program("goes", "vnet", "show", "ip", "fib")
}

func checkOspfRunning(t *testing.T) {
	assert := test.Assert{t}

	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range ospf_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, ospf_config, cmd)
		assert.Nil(err)
		assert.Match(out, ".*ospfd.*")
		assert.Match(out, ".*zebra.*")
	}
}

func checkOspfNeighbors(t *testing.T) {
	assert := test.Assert{t}

	time.Sleep(60 * time.Second) // give ospf time to converge
	cmd := []string{"vtysh", "-c", "show ip ospf neig"}
	for _, r := range ospf_config.Routers {
		out, err := docker.ExecCmd(t, r.Hostname, ospf_config, cmd)
		assert.Nil(err)
		assert.Match(out, "192.168.*")
	}
}

func checkOspfLearnedRoute(t *testing.T) {
	assert := test.Assert{t}

	cmd := []string{"ip", "route", "show", "192.168.222.0/24"}
	out, err := docker.ExecCmd(t, "R1", ospf_config, cmd)
	assert.Nil(err)
	assert.Match(out, "192.168.222.0/24.*")
}

func checkOspfConnectivityLearned(t *testing.T) {
	assert := test.Assert{t}

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2")
	assert.Program("goes", "vnet", "show", "ip", "fib")
}
