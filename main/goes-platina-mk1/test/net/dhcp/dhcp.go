// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dhcp

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
	Name: "dhcp",
	Tests: test.Tests{
		&docket{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/net/dhcp/conf.yaml.tmpl",
			},
		},
		&docket{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/net/dhcp/vlan/conf.yaml.tmpl",
			},
		},
	},
}

func (d *docket) Run(t *testing.T) {
	d.UTS(t, []test.UnitTest{
		{"connectivity", d.checkConnectivity},
		{"server", d.checkServer},
		{"client", d.checkClient},
		{"connectivity2", d.checkConnectivity2},
		{"vlan-tag", d.checkVlanTag},
	})
}

func (d *docket) checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		host   string
		target string
	}{
		{"R1", "192.168.120.10"},
		{"R2", "192.168.120.5"},
	} {
		out, err := d.ExecCmd(t, x.host, "ping", "-c3", x.target)
		assert.Nil(err)
		assert.Match(out, "[1-3] packets received")
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.host)
	}
}

func (d *docket) checkServer(t *testing.T) {
	assert := test.Assert{t}

	t.Logf("Checking dhcp server on %v", "R2")
	out, err := d.ExecCmd(t, "R2", "ps", "ax")
	assert.Nil(err)
	assert.Match(out, ".*dhcpd.*")
}

func (d *docket) checkClient(t *testing.T) {
	assert := test.Assert{t}

	r, err := docker.FindHost(d.Config, "R1")
	intf := r.Intfs[0]

	// remove existing IP address
	_, err = d.ExecCmd(t, "R1",
		"ip", "address", "delete", "192.168.120.5", "dev", intf.Name)
	assert.Nil(err)

	t.Log("Verify ping fails")
	_, err = d.ExecCmd(t, "R1", "ping", "-c1", "192.168.120.10")
	assert.NonNil(err)

	t.Log("Request dhcp address")
	out, err := d.ExecCmd(t, "R1", "dhclient", "-4", "-v", intf.Name)
	assert.Nil(err)
	assert.Match(out, "bound to")
}

func (d *docket) checkConnectivity2(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check connectivity with dhcp address")
	out, err := d.ExecCmd(t, "R1", "ping", "-c3", "192.168.120.10")
	assert.Nil(err)
	assert.Match(out, "[1-3] packets received")
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R1")
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R2")
}

func (d *docket) checkVlanTag(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check for invalid vlan tag") // issue #92

	r1, err := docker.FindHost(d.Config, "R1")
	r1Intf := r1.Intfs[0]

	// remove existing IP address
	_, err = d.ExecCmd(t, "R1",
		"ip", "address", "flush", "dev", r1Intf.Name)
	assert.Nil(err)

	r2, err := docker.FindHost(d.Config, "R2")
	r2Intf := r2.Intfs[0]

	done := make(chan bool, 1)

	go func(done chan bool) {
		out, err := d.ExecCmd(t, "R2",
			"timeout", "10",
			"tcpdump", "-c1", "-nvvvei", r2Intf.Name, "port", "67")
		assert.Nil(err)
		match, err := regexp.MatchString("vlan 0", out)
		assert.Nil(err)
		if match {
			t.Error("Invalid vlan 0 tag found")
		}
		done <- true
	}(done)

	time.Sleep(1 * time.Second)
	_, err = d.ExecCmd(t, "R1", "dhclient", "-4", "-v", r1Intf.Name)
	assert.Nil(err)
	<-done
}
