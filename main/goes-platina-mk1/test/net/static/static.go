// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package static

import (
	"regexp"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

var config *docker.Config

func Test(t *testing.T, source []byte) {
	var err error
	config, err = docker.LaunchContainers(t, source)
	if err != nil {
		t.Fatalf("Error launchContainers: %v", err)
	}
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

	for _, x := range []struct {
		hostname string
		target   string
	}{
		{"CA-1", "10.1.0.2"},
		{"RA-1", "10.1.0.1"},
		{"RA-1", "10.2.0.3"},
		{"RA-2", "10.2.0.2"},
		{"RA-2", "10.3.0.4"},
		{"CA-2", "10.3.0.3"},
	} {
		t.Logf("ping from %v to %v", x.hostname, x.target)
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
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
		assert.True(regexp.MustCompile(".*zebra.*").MatchString(out))
	}
}

func checkRoutes(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range config.Routers {

		t.Logf("check for default route in container RIB %v",
			r.Hostname)
		cmd := []string{"vtysh", "-c", "show ip route"}
		out, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "S>\\* 0.0.0.0/0")

		t.Logf("check for default route in container FIB %v", r.Hostname)
		cmd = []string{"ip", "route", "show"}
		out, err = docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "default")

		t.Logf("check for default route in goes fib %v", r.Hostname)
		assert.Program(regexp.MustCompile("0.0.0.0/0"),
			test.Self{},
			"vnet", "show", "ip", "fib", "table", r.Hostname)
	}
}

func checkInterConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		target   string
	}{
		{"CA-1", "10.3.0.4"},
		{"CA-2", "10.1.0.1"},
	} {
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}
