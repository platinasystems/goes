// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package slice

import (
	"regexp"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

var config *docker.Config

func Test(t *testing.T, source []byte) {
	config = docker.LaunchContainers(t, source)
	defer docker.TearDownContainers(t, config)

	test.Suite{
		{"connectivity", checkConnectivity},
		{"frr", checkFrr},
		{"routes", checkRoutes},
		{"inter-connectivity", checkInterConnectivity},
	}.Run(t)
}

func checkConnectivity(t *testing.T) {
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

func checkFrr(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)
	cmd := []string{"ps", "ax"}
	for _, r := range config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		if err != nil {
			t.Logf("DockerExecCmd failed: %v", err)
			t.Fail()
			return
		}
		assert.True(regexp.MustCompile(".*ospfd.*").MatchString(out))
		assert.True(regexp.MustCompile(".*zebra.*").MatchString(out))
	}
}

func checkRoutes(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(60 * time.Second)

	cmd := []string{"ip", "route", "show"}
	out, err := docker.ExecCmd(t, "CA-1", config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.Match(out, "10.3.0.0/24")

	out, err = docker.ExecCmd(t, "CA-2", config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.Match(out, "10.1.0.0/24")

	out, err = docker.ExecCmd(t, "CB-1", config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.Match(out, "10.3.0.0/24")

	out, err = docker.ExecCmd(t, "CB-2", config, cmd)
	if err != nil {
		t.Logf("DockerExecCmd failed: %v", err)
		t.Fail()
		return
	}
	assert.Match(out, "10.1.0.0/24")
}

func checkInterConnectivity(t *testing.T) {
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
