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
		{"isolation", checkIsolation},
		{"stress", checkStress},
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
		{"CB-1", "10.1.0.2"},
		{"RB-1", "10.1.0.1"},
		{"RB-1", "10.2.0.3"},
		{"RB-2", "10.2.0.2"},
		{"RB-2", "10.3.0.4"},
		{"CB-2", "10.3.0.3"},
	} {
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
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
		assert.True(regexp.MustCompile(".*ospfd.*").MatchString(out))
		assert.True(regexp.MustCompile(".*zebra.*").MatchString(out))
	}
}

func checkRoutes(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		route    string
	}{
		{"CA-1", "10.3.0.0/24"},
		{"CA-2", "10.1.0.0/24"},
		{"CB-1", "10.3.0.0/24"},
		{"CB-2", "10.1.0.0/24"},
	} {
		found := false
		cmd := []string{"ip", "route", "show", x.route}
		timeout := 120
		for i := timeout; i > 0; i-- {
			out, err := docker.ExecCmd(t, x.hostname, config, cmd)
			assert.Nil(err)
			if !assert.MatchNonFatal(out, x.route) {
				time.Sleep(1 * time.Second)
			} else {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("No ospf route for %v: %v", x.hostname, x.route)
		}
	}
}

func checkInterConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		hostname string
		target   string
	}{
		{"CA-1", "10.3.0.4"}, // In slice A ping from CA-1 to CA-2
		{"CB-1", "10.3.0.4"}, // In slice B ping from CB-1 to CB-2
		{"CA-2", "10.1.0.1"}, // In slice A ping from CA-2 to CA-1
		{"CB-2", "10.1.0.1"}, // In slice B ping from CB-2 to CB-1

	} {
		cmd := []string{"ping", "-c3", x.target}
		out, err := docker.ExecCmd(t, x.hostname, config, cmd)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}

func checkIsolation(t *testing.T) {
	assert := test.Assert{t}

	// break slice B connectivity does not affect slice A
	r, err := docker.FindHost(config, "RB-2")
	assert.Nil(err)

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
	}
	// how do I do an anti match???
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "RB-2")

	t.Log("Verify that slice B is broken")
	cmd := []string{"ping", "-c1", "10.3.0.4"}
	_, err = docker.ExecCmd(t, "CB-1", config, cmd)
	assert.NonNil(err)

	t.Log("Verify that slice A is not affected")
	cmd = []string{"ping", "-c1", "10.3.0.4"}
	_, err = docker.ExecCmd(t, "CA-1", config, cmd)
	assert.Nil(err)
	assert.Program(regexp.MustCompile("10.3.0.0/24"),
		test.Self{},
		"vnet", "show", "ip", "fib", "table", "RA-2")

	// bring RB-2 interfaces back up
	for _, i := range r.Intfs {
		var intf string
		if i.Vlan != "" {
			intf = i.Name + "." + i.Vlan
		} else {
			intf = i.Name
		}
		cmd := []string{"ip", "link", "set", "up", intf}
		_, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
	}

	// break slice A connectivity does not affect slice B
	r, err = docker.FindHost(config, "RA-2")
	assert.Nil(err)

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
	}
	// how do I do an anti match???
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "RA-2")

	t.Log("Verify that slice A is broken")
	cmd = []string{"ping", "-c1", "10.3.0.4"}
	_, err = docker.ExecCmd(t, "CA-1", config, cmd)
	assert.NonNil(err)

	ok := false
	t.Log("Verify that slice B is not affected")
	timeout := 120
	for i := timeout; i > 0; i-- {
		cmd = []string{"ping", "-c1", "10.3.0.4"}
		out, _ := docker.ExecCmd(t, "CB-1", config, cmd)
		if !assert.MatchNonFatal(out, "1 packets received") {
			time.Sleep(1 * time.Second)
		} else {
			ok = true
			break
		}
	}
	if !ok {
		t.Error("Slice B ping failed")
	}
	assert.Program(regexp.MustCompile("10.3.0.0/24"),
		test.Self{},
		"vnet", "show", "ip", "fib", "table", "RB-2")

	// bring RA-1 interfaces back up
	for _, i := range r.Intfs {
		var intf string
		if i.Vlan != "" {
			intf = i.Name + "." + i.Vlan
		} else {
			intf = i.Name
		}
		cmd := []string{"ip", "link", "set", "up", intf}
		_, err := docker.ExecCmd(t, r.Hostname, config, cmd)
		assert.Nil(err)
	}

}

func checkStress(t *testing.T) {
	assert := test.Assert{t}

	t.Skip() // for now even 1 second hping3 --faster hangs tx

	t.Log("stress with hping3")

	duration := []string{"1", "10", "30", "60"}

	ok := false
	timeout := 120
	for i := timeout; i > 0; i-- {
		cmd := []string{"ping", "-c1", "10.3.0.4"}
		out, _ := docker.ExecCmd(t, "CB-1", config, cmd)
		if !assert.MatchNonFatal(out, "1 packets received") {
			time.Sleep(1 * time.Second)
		} else {
			ok = true
			t.Log("ping ok before stress")
			break
		}
	}
	if !ok {
		t.Error("ping failing before stress test")
	}

	for _, to := range duration {
		t.Logf("stress for %v", to)
		cmd := []string{"timeout", to,
			"hping3", "--icmp", "--faster", "-q", "10.3.0.4"}
		_, err := docker.ExecCmd(t, "CB-1", config, cmd)
		// t.Logf("hping3 duration %v [%v] [%v]", to, out, err)
		t.Log("verfy can still ping neighbor")
		cmd = []string{"ping", "-c1", "10.1.0.2"}
		_, err = docker.ExecCmd(t, "CB-1", config, cmd)
		// t.Logf("ping check [%v] [%v]", out, err)
		assert.Nil(err)
	}
}
