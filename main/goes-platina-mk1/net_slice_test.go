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

var slice_config *docker.Config

func Slice(t *testing.T, confFile string) {
	slice_config = docker.LaunchContainers(t, confFile)
	defer docker.TearDownContainers(t, slice_config)

	test.Suite{
		{"l2", sliceL2Connectivity},
		{"frr", checkSliceOspfRunning},
		{"routes", checkSliceOspfLearnedRoute},
		{"ping-learned", checkSliceOspfConnectivityLearned},
	}.Run(t)
}

func sliceL2Connectivity(t *testing.T) {
	assert := test.Assert{t}

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CA-1",
		"ping", "-c1", "10.1.0.2")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RA-1",
		"ping", "-c1", "10.1.0.1")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RA-1",
		"ping", "-c1", "10.2.0.3")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RA-2",
		"ping", "-c1", "10.2.0.2")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RA-2",
		"ping", "-c1", "10.3.0.4")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CA-2",
		"ping", "-c1", "10.3.0.3")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CB-1",
		"ping", "-c1", "10.1.0.2")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RB-1",
		"ping", "-c1", "10.1.0.1")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RB-1",
		"ping", "-c1", "10.2.0.3")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RB-2",
		"ping", "-c1", "10.2.0.2")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "RB-2",
		"ping", "-c1", "10.3.0.4")

	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CB-2",
		"ping", "-c1", "10.3.0.3")

	assert.Program("goes", "vnet", "show", "ip", "fib")
}

func checkSliceOspfRunning(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range slice_config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, slice_config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		assert.True(regexp.MustCompile(".*ospfd.*").MatchString(out))
		assert.True(regexp.MustCompile(".*zebra.*").MatchString(out))
	}
}

func checkSliceOspfLearnedRoute(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(60 * time.Second)

	cmd := []string{"ip", "route", "show"}
	out, err := docker.ExecCmd(t, "CA-1", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.True(out == "10.3.0.0/24")

	out, err = docker.ExecCmd(t, "CA-2", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.True(out == "10.1.0.0/24")

	out, err = docker.ExecCmd(t, "CB-1", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.True(out == "10.3.0.0/24")

	out, err = docker.ExecCmd(t, "CB-2", slice_config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.True(out == "10.1.0.0/24")
}

func checkSliceOspfConnectivityLearned(t *testing.T) {
	assert := test.Assert{t}

	// In slice A ping from CA-1 to CA-2
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CA-1",
		"ping", "-c1", "10.3.0.4")

	// In slice B ping from CB-1 to CB-2
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CB-1",
		"ping", "-c1", "10.3.0.4")

	// In slice A ping from CA-2 to CA-1
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CA-2",
		"ping", "-c1", "10.1.0.1")

	// In slice B ping from CB-2 to CB-1
	assert.Program(regexp.MustCompile("1 received"),
		"goes", "ip", "netns", "exec", "CB-2",
		"ping", "-c1", "10.1.0.1")

	assert.Program("goes", "vnet", "show", "ip", "fib")
}
