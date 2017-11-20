// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package isis

import (
	"regexp"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

var config *docker.Config

func Test(t *testing.T, yaml []byte) {

	config = docker.LaunchContainers(t, yaml)
	defer docker.TearDownContainers(t, config)

	test.Suite{
		{"connectivity", checkConnectivity},
		{"frr", checkFrr},
		{"neighbors", checkNeighbors},
		{"routes", checkRoutes},
		{"inter-connectivity", checkInterConnectivity},
	}.Run(t)
}

func checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	assert.Program(regexp.MustCompile("1 received"),
		test.Self{}, "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.120.10")

	assert.Program(regexp.MustCompile("1 received"),
		test.Self{}, "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.150.4")

	assert.Program(regexp.MustCompile("1 received"),
		test.Self{}, "ip", "netns", "exec", "R2",
		"ping", "-c1", "192.168.222.2")

	assert.Program(regexp.MustCompile("1 received"),
		test.Self{}, "ip", "netns", "exec", "R3",
		"ping", "-c1", "192.168.111.4")
}

func checkFrr(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)

	cmd := []string{"ps", "ax"}
	for _, r := range config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, ".*isisd.*")
		assert.Match(out, ".*zebra.*")
	}
}

func checkNeighbors(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(60 * time.Second) // give time to converge

	cmd := []string{"vtysh", "-c", "show isis interface"}
	for _, r := range config.Routers {
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "Area R.*")
	}

	cmd = []string{"vtysh", "-c", "show isis summary"}
	for _, r := range config.Routers {
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, ".*Net: 47.0023.*")
	}
}

func checkRoutes(t *testing.T) {
	assert := test.Assert{t}
	cmd := []string{"ip", "route", "show", "192.168.222.0/24"}
	out, err := docker.ExecCmd(t, "R1", config, cmd)
	assert.Nil(err)
	test.Pause()
	assert.Match(out, "192.168.222.0/24.*")
}

func checkInterConnectivity(t *testing.T) {
	assert := test.Assert{t}
	assert.Program(regexp.MustCompile("1 received"),
		test.Self{}, "ip", "netns", "exec", "R1",
		"ping", "-c1", "192.168.222.2")
	assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
}
