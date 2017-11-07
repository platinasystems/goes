// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
)

var ospf_config *Config

func FrrOSPF(t *testing.T, confFile string) {

	err := CheckDocker(t)
	if err != nil {
		t.Fatal(err)
	}

	assert := Assert{t}
	assert.YoureRoot()
	defer assert.Program(nil,
		"goes", "redisd",
	).Quit(10 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "redis.ready", "true", "10",
	).Ok()
	defer assert.Program(nil,
		"goes", "vnetd",
	).Gdb().Quit(30 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "vnet.ready", "true", "30",
	).Ok()

	ospf_config = LaunchContainers(t, confFile)

	runOspfTestCases(t)

	TearDownContainers(t, ospf_config)
}

func runOspfTestCases(t *testing.T) {
	t.Run("check L2 connectivity", ospfCheckL2Connectivity)
	time.Sleep(1 * time.Second)
	t.Run("check FRR running", checkOspfRunning)
	time.Sleep(60 * time.Second) // give ospf time to converge
	t.Run("check ospf neighbors", checkOspfNeighbors)
	t.Run("check learned route", checkOspfLearnedRoute)
}

func ospfCheckL2Connectivity(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.120.10",
	).Output(Match("1 received"))

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.150.4",
	).Output(Match("1 received"))

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R2",
		"ping", "-c1", "192.168.222.2",
	).Output(Match("1 received"))

	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R3",
		"ping", "-c1", "192.168.111.4",
	).Output(Match("1 received"))
}

func checkOspfRunning(t *testing.T) {

	cmd := []string{"ps", "ax"}
	for _, r := range ospf_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := DockerExecCmd(t, r.Hostname, ospf_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output(Match("ospfd"))
		Assert{t}.Program(nil,
			"echo", out,
		).Output(Match("zebra"))
	}
}

func checkOspfNeighbors(t *testing.T) {

	cmd := []string{"vtysh", "-c", "show ip ospf neig"}
	for _, r := range ospf_config.Routers {
		out, err := DockerExecCmd(t, r.Hostname, ospf_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output(Match("192.168."))
	}
}

func checkOspfLearnedRoute(t *testing.T) {

	cmd := []string{"ip", "route", "show"}
	out, err := DockerExecCmd(t, "R1", ospf_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output(Match("192.168.222.0/24"))
}

func checkOspfConnectivityLearned(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2",
	).Output(Match("1 received"))

}
