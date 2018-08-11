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

type dhcp struct {
	docker.Docket
}

var Suite = test.Suite{
	Name: "dhcp",
	Tests: test.Tests{
		&dhcp{
			docker.Docket{
				Name: "eth",
				Tmpl: "testdata/net/dhcp/conf.yaml.tmpl",
			},
		},
		&dhcp{
			docker.Docket{
				Name: "vlan",
				Tmpl: "testdata/net/dhcp/vlan/conf.yaml.tmpl",
			},
		},
	},
}

func (dhcp *dhcp) Test(t *testing.T) {
	dhcp.Docket.Tests = test.Tests{
		&test.Unit{"check connectivity", dhcp.checkConnectivity},
		&test.Unit{"check dhcp server", dhcp.checkServer},
		&test.Unit{"check dhcp client", dhcp.checkClient},
		&test.Unit{"check connectivity again",
			dhcp.checkConnectivity2},
		&test.Unit{"check vlan tag", dhcp.checkVlanTag},
	}
	dhcp.Docket.Test(t)
}

func (dhcp *dhcp) checkConnectivity(t *testing.T) {
	assert := test.Assert{t}

	for _, x := range []struct {
		host   string
		target string
	}{
		{"R1", "192.168.120.10"},
		{"R2", "192.168.120.5"},
	} {
		err := dhcp.PingCmd(t, x.host, x.target)
		assert.Nil(err)
		assert.Program(test.Self{},
			"vnet", "show", "ip", "fib", "table", x.host)
	}
}

func (dhcp *dhcp) checkServer(t *testing.T) {
	assert := test.Assert{t}

	t.Logf("Checking dhcp server on %v", "R2")
	out, err := dhcp.ExecCmd(t, "R2", "ps", "ax")
	assert.Nil(err)
	assert.Match(out, ".*dhcpd.*")
}

func (dhcp *dhcp) checkClient(t *testing.T) {
	assert := test.Assert{t}

	r, err := docker.FindHost(dhcp.Config, "R1")
	intf := r.Intfs[0]

	// remove existing IP address
	_, err = dhcp.ExecCmd(t, "R1",
		"ip", "address", "delete", "192.168.120.5", "dev", intf.Name)
	assert.Nil(err)

	t.Log("Verify ping fails")
	_, err = dhcp.ExecCmd(t, "R1", "ping", "-c1", "192.168.120.10")
	assert.NonNil(err)

	t.Log("Request dhcp address")
	out, err := dhcp.ExecCmd(t, "R1", "dhclient", "-4", "-v", intf.Name)
	assert.Nil(err)
	assert.Match(out, "bound to")
}

func (dhcp *dhcp) checkConnectivity2(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check connectivity with dhcp address")
	err := dhcp.PingCmd(t, "R1", "192.168.120.10")
	assert.Nil(err)
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R1")
	assert.Program(test.Self{},
		"vnet", "show", "ip", "fib", "table", "R2")
}

func (dhcp *dhcp) checkVlanTag(t *testing.T) {
	assert := test.Assert{t}

	t.Log("Check for invalid vlan tag") // issue #92

	r1, err := docker.FindHost(dhcp.Config, "R1")
	r1Intf := r1.Intfs[0]

	// remove existing IP address
	_, err = dhcp.ExecCmd(t, "R1",
		"ip", "address", "flush", "dev", r1Intf.Name)
	assert.Nil(err)

	r2, err := docker.FindHost(dhcp.Config, "R2")
	r2Intf := r2.Intfs[0]

	done := make(chan bool, 1)

	go func(done chan bool) {
		out, err := dhcp.ExecCmd(t, "R2",
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
	_, err = dhcp.ExecCmd(t, "R1", "dhclient", "-4", "-v", r1Intf.Name)
	assert.Nil(err)
	<-done
}
