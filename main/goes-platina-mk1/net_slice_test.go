// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
)

var slice_config *Config

func Slice(t *testing.T, confFile string) {

	slice_config = LaunchContainers(t, confFile)
	defer TearDownContainers(t, slice_config)

	Suite{
		{"l2", sliceL2Connectivity},
		{"frr", checkSliceOspfRunning},
		{"routes", checkSliceOspfLearnedRoute},
		{"ping-learned", checkSliceOspfConnectivityLearned},
	}.Run(t)
}

func sliceL2Connectivity(t *testing.T) {
	assert := Assert{t}
	assert.Program(nil,
		"goes", "ip", "netns", "exec", "CA-1",
		"ping", "-c1", "10.1.0.2",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RA-1",
		"ping", "-c1", "10.1.0.1",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RA-1",
		"ping", "-c1", "10.2.0.3",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RA-2",
		"ping", "-c1", "10.2.0.2",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RA-2",
		"ping", "-c1", "10.3.0.4",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "CA-2",
		"ping", "-c1", "10.3.0.3",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "CB-1",
		"ping", "-c1", "10.1.0.2",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RB-1",
		"ping", "-c1", "10.1.0.1",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RB-1",
		"ping", "-c1", "10.2.0.3",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RB-2",
		"ping", "-c1", "10.2.0.2",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "RB-2",
		"ping", "-c1", "10.3.0.4",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "ip", "netns", "exec", "CB-2",
		"ping", "-c1", "10.3.0.3",
	).Output("/1 received/").Done()

	assert.Program(nil,
		"goes", "vnet", "show", "ip", "fib",
	).Ok().Done()
}

func checkSliceOspfRunning(t *testing.T) {
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range slice_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := DockerExecCmd(t, r.Hostname, slice_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/ospfd/").Done()
		Assert{t}.Program(nil,
			"echo", out,
		).Output("/zebra/").Done()
	}
}

func checkSliceOspfLearnedRoute(t *testing.T) {
	time.Sleep(60 * time.Second)

	cmd := []string{"ip", "route", "show"}
	out, err := DockerExecCmd(t, "CA-1", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/10.3.0.0/24/").Done()

	out, err = DockerExecCmd(t, "CA-2", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/10.1.0.0/24/").Done()

	out, err = DockerExecCmd(t, "CB-1", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/10.3.0.0/24/").Done()

	out, err = DockerExecCmd(t, "CB-2", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	Assert{t}.Program(nil,
		"echo", out,
	).Output("/10.1.0.0/24/").Done()
}

func checkSliceOspfConnectivityLearned(t *testing.T) {
	// In slice A ping from CA-1 to CA-2
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "CA-1",
		"ping", "-c1", "10.3.0.4",
	).Output("/1 received/").Done()

	// In slice B ping from CB-1 to CB-2
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "CB-1",
		"ping", "-c1", "10.3.0.4",
	).Output("/1 received/").Done()

	// In slice A ping from CA-2 to CA-1
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "CA-2",
		"ping", "-c1", "10.1.0.1",
	).Output("/1 received/").Done()

	// In slice B ping from CB-2 to CB-1
	Assert{t}.Program(nil,
		"goes", "ip", "netns", "exec", "CB-2",
		"ping", "-c1", "10.1.0.1",
	).Output("/1 received/").Done()

	Assert{t}.Program(nil,
		"goes", "vnet", "show", "ip", "fib",
	).Ok().Done()
}
