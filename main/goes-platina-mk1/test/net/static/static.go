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
	"github.com/platinasystems/go/main/goes-platina-mk1/test/conf"
)

var config *docker.Config

var Suite = test.Suite{
	{"eth", func(t *testing.T) {
		subtest(t, conf.New(t, "testdata/net/static/eth/conf.yaml.tmpl"))
	}},
	{"vlan", func(t *testing.T) {
		subtest(t, conf.New(t, "testdata/net/static/vlan/conf.yaml.tmpl"))
	}},
}.Run

func subtest(t *testing.T, yaml []byte) {
	var err error
	config, err = docker.LaunchContainers(t, yaml)
	if err != nil {
		t.Fatalf("Error launchContainers: %v", err)
	}
	defer docker.TearDownContainers(t, config)

	test.Suite{
		{"connectivity", checkConnectivity},
		{"frr", checkFrr},
		{"routes", checkRoutes},
		{"inter-connectivity", checkInterConnectivity},
		{"flap", checkFlap},
		{"inter-connectivity2", checkInterConnectivity2},
		{"punt-stress", checkPuntStress},
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
		{"RA-1", "192.168.0.1"},
		{"RA-2", "10.2.0.2"},
		{"RA-2", "10.3.0.4"},
		{"RA-2", "192.168.0.2"},
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
		{"CA-1", "192.168.0.2"},
		{"CA-2", "10.1.0.1"},
		{"CA-2", "192.168.0.1"},
	} {
		t.Logf("ping from %v to %v", x.hostname, x.target)
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}

func checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range config.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			cmd := []string{"ip", "link", "set", "down", intf}
			_, err := docker.ExecCmd(t, r.Hostname, config, cmd)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			cmd = []string{"ip", "link", "set", "up", intf}
			_, err = docker.ExecCmd(t, r.Hostname, config, cmd)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}

func checkInterConnectivity2(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		target   string
	}{
		{"CA-1", "10.1.0.2"},
		{"RA-1", "10.1.0.1"},
		{"RA-1", "10.2.0.3"},
		{"RA-1", "192.168.0.1"},
		{"RA-2", "10.2.0.2"},
		{"RA-2", "10.3.0.4"},
		{"RA-2", "192.168.0.2"},
		{"CA-2", "10.3.0.3"},
		{"CA-1", "10.3.0.4"},
		{"CA-1", "192.168.0.2"},
		{"CA-2", "10.1.0.1"},
		{"CA-2", "192.168.0.1"},
	} {
		t.Logf("ping from %v to %v", x.hostname, x.target)
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}

func checkPuntStress(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check punt stress with iperf3")

	done := make(chan bool, 1)

	go func(done chan bool) {
		cmd := []string{"timeout", "15", "iperf3", "-s"}
		docker.ExecCmd(t, "CA-2", config, cmd)
		done <- true
	}(done)

	time.Sleep(1 * time.Second)
	cmd := []string{"iperf3", "-c", "10.3.0.4"}
	out, err := docker.ExecCmd(t, "CA-1", config, cmd)

	r, err := regexp.Compile(`([0-9\.]+)\s+Gbits/sec\s+receiver`)
	assert.Nil(err)
	result := r.FindStringSubmatch(out)
	if len(result) == 2 {
		t.Logf("iperf3 - %v Gbits/sec", result[1])
	} else {
		t.Logf("iperf3 regex failed to find rate [%v]", out)
	}
	<-done
}
