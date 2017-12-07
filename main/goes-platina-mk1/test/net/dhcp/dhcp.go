// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dhcp

import (
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
)

var config *docker.Config

func Test(t *testing.T, yaml []byte) {
	var err error
	config, err = docker.LaunchContainers(t, yaml)
	if err != nil {
		t.Fatalf("Error launchContainers: %v", err)
	}
	defer docker.TearDownContainers(t, config)

	test.Suite{
		{"connectivity", checkConnectivity},
		{"server", checkServer},
		{"client", checkClient},
		{"connectivity2", checkConnectivity2},
	}.Run(t)
}

func checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		host   string
		target string
	}{
		{"R1", "192.168.120.10"},
		{"R2", "192.168.120.5"},
	} {
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.host, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.host)
	}
}

func checkServer(t *testing.T) {
	assert := test.Assert{t}

	cmd := []string{"ps", "ax"}
	t.Logf("Checking dhcp server on %v", "R2")
	out, err := docker.ExecCmd(t, "R2", config, cmd)
	assert.Nil(err)
	assert.Match(out, ".*dhcpd.*")
}

func checkClient(t *testing.T) {
	assert := test.Assert{t}

	r, err := docker.FindHost(config, "R1")
	intf := r.Intfs[0]

	// remove existing IP address
	cmd := []string{"ip", "address", "delete", "192.168.120.5", "dev",
		intf.Name}
	_, err = docker.ExecCmd(t, "R1", config, cmd)
	assert.Nil(err)

	t.Log("Verify ping fails")
	cmd = []string{"ping", "-c1", "192.168.120.10"}
	_, err = docker.ExecCmd(t, "R1", config, cmd)
	assert.NonNil(err)

	t.Log("Request dhcp address")
	cmd = []string{"dhclient", "-4", "-v", intf.Name}
	out, err := docker.ExecCmd(t, "R1", config, cmd)
	assert.Nil(err)
	assert.Match(out, "bound to")
}

func checkConnectivity2(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check connectivity with dhcp address")
	cmd := []string{"ping", "-c3", "192.168.120.10"}
	out, err := docker.ExecCmd(t, "R1", config, cmd)
	assert.Nil(err)
	assert.Match(out, "[1-3] packets received")
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R1")
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R2")
}
