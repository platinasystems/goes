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

type docket struct {
	docker.Docket
}

var Suite = test.Suite{
	Name: "static",
	Tests: test.Tests{
		&docket{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/net/static/conf.yaml.tmpl",
			},
		},
		&docket{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/net/static/conf.yaml.tmpl",
			},
		},
	},
}

func (d *docket) Run(t *testing.T) {
	d.UTS(t, []test.UnitTest{
		test.UnitTest{"connectivity", d.checkConnectivity},
		test.UnitTest{"frr", d.checkFrr},
		test.UnitTest{"routes", d.checkRoutes},
		test.UnitTest{"inter-connectivity", d.checkInterConnectivity},
		test.UnitTest{"flap", d.checkFlap},
		test.UnitTest{"inter-connectivity2", d.checkInterConnectivity2},
		test.UnitTest{"punt-stress", d.checkPuntStress},
	})
}

func (d *docket) checkConnectivity(t *testing.T) {
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
		out, err := d.ExecCmd(t, x.hostname, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
	}
}

func (d *docket) checkFrr(t *testing.T) {
	assert := test.Assert{t}
	time.Sleep(1 * time.Second)

	for _, r := range d.Config.Routers {
		t.Logf("Checking FRR on %v", r.Hostname)
		out, err := d.ExecCmd(t, r.Hostname, "ps", "ax")
		assert.Nil(err)
		assert.True(regexp.MustCompile(".*zebra.*").MatchString(out))
	}
}

func (d *docket) checkRoutes(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range d.Config.Routers {

		t.Logf("check for default route in container RIB %v",
			r.Hostname)
		out, err := d.ExecCmd(t, r.Hostname, "vtysh", "-c",
			"show ip route")
		assert.Nil(err)
		assert.Match(out, "S>\\* 0.0.0.0/0")

		t.Logf("check for default route in container FIB %v", r.Hostname)
		out, err = d.ExecCmd(t, r.Hostname, "ip", "route", "show")
		assert.Nil(err)
		assert.Match(out, "default")

		t.Logf("check for default route in goes fib %v", r.Hostname)
		assert.Program(regexp.MustCompile("0.0.0.0/0"),
			test.Self{},
			"vnet", "show", "ip", "fib", "table", r.Hostname)
	}
}

func (d *docket) checkInterConnectivity(t *testing.T) {
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
		out, err := d.ExecCmd(t, x.hostname, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}

func (d *docket) checkFlap(t *testing.T) {
	assert := test.Assert{t}

	for _, r := range d.Config.Routers {
		for _, i := range r.Intfs {
			var intf string
			if i.Vlan != "" {
				intf = i.Name + "." + i.Vlan
			} else {
				intf = i.Name
			}
			_, err := d.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "down", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			_, err = d.ExecCmd(t, r.Hostname,
				"ip", "link", "set", "up", intf)
			assert.Nil(err)
			time.Sleep(1 * time.Second)
			assert.Program(test.Self{}, "vnet", "show", "ip", "fib")
		}
	}
}

func (d *docket) checkInterConnectivity2(t *testing.T) {
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
		out, err := d.ExecCmd(t, x.hostname, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.hostname)
	}
}

func (d *docket) checkPuntStress(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check punt stress with iperf3")

	done := make(chan bool, 1)

	go func(done chan bool) {
		d.ExecCmd(t, "CA-2", "timeout", "15", "iperf3", "-s")
		done <- true
	}(done)

	time.Sleep(1 * time.Second)
	out, err := d.ExecCmd(t, "CA-1", "iperf3", "-c", "10.3.0.4")

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
