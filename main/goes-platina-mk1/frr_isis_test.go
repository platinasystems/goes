// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
)

var isis_config *Config

func FrrISIS(t *testing.T, confFile string) {

	isis_config = LaunchContainers(t, confFile)
	defer TearDownContainers(t, isis_config)

	Suite{
		{"l2", isisCheckL2Connectivity},
		{"frr", checkIsisRunning},
		{"neighbors", checkIsisNeighbors},
		{"routes", checkIsisLearnedRoute},
		{"ping-learned", checkIsisConnectivityLearned},
	}.Run(t)
}

func isisCheckL2Connectivity(t *testing.T) {
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
}

func checkIsisRunning(t *testing.T) {
	time.Sleep(1 * time.Second)

	cmd := []string{"ps", "ax"}
	for _, r := range isis_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := DockerExecCmd(t, r.Hostname, isis_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/isisd/").Done()
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/zebra/").Done()
	}
}

func checkIsisNeighbors(t *testing.T) {
	time.Sleep(60 * time.Second) // give time to converge

	cmd := []string{"vtysh", "-c", "show isis interface"}
	for _, r := range isis_config.Routers {
		out, err := DockerExecCmd(t, r.Hostname, isis_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/Area R/").Done()
	}

	cmd = []string{"vtysh", "-c", "show isis summary"}
	for _, r := range isis_config.Routers {
		out, err := DockerExecCmd(t, r.Hostname, isis_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/Net: 47.0023/").Done()
	}
}

func checkIsisLearnedRoute(t *testing.T) {
	cmd := []string{"ip", "route", "show"}
	out, err := DockerExecCmd(t, "R1", isis_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/192.168.222.0/24/").Done()
}

func checkIsisConnectivityLearned(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2",
	).Output("/1 received/").Done()
	Assert{t}.Program(nil,
		"goes", "vnet", "show", "ip", "fib",
	).Ok().Done()
}
