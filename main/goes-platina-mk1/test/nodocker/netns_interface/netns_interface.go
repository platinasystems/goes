// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns_interface

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

type route struct {
	prefix string
	gw     string
}

type config struct {
	netns   string
	ifname  string
	ifa     string
	routes  []route
	remotes []string
}

type nodocker map[string]*config

var Suite = test.Suite{
	Name: "netns_interface",
	Tests: test.Tests{
		&nodocker{
			"net0port0": &config{
				netns:   "R1",
				ifa:     "10.0.0.1/24",
				remotes: []string{"10.0.0.2"},
			},
			"net0port1": &config{
				netns:   "R2",
				ifa:     "10.0.0.2/24",
				remotes: []string{"10.0.0.1"},
			},
		},
	},
}

func (nodocker) String() string { return "eth" }

func (m nodocker) Test(t *testing.T) {
	assert := test.Assert{t}
	for k, c := range m {
		ifname := netport.PortByNetPort[k]
		c.ifname = ifname
		ns := c.netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err != nil {
			assert.Program(test.Self{},
				"ip", "netns", "add", ns)
		}
		assert.Program(test.Self{},
			"ip", "link", "set", ifname, "up", "netns", ns)
		assert.Program(test.Self{},
			"ip", "netns", "exec", ns, test.Self{},
			"ip", "address", "add", c.ifa, "dev", ifname)
	}
	for _, c := range m {
		assert.Nil(test.Carrier(c.netns, c.ifname))
	}
	test.Tests{
		&test.Unit{"ping neighbor", m.pingNeighbor},
		&test.Unit{"check neighbor", m.checkNeighbor},
		&test.Unit{"delete netns", m.delNetns},
		&test.Unit{"check no neighbor", m.checkNoNeighbor},
	}.Test(t)
}

func (m nodocker) pingNeighbor(t *testing.T) {
	assert := test.Assert{t}
	for _, c := range m {
		for _, n := range c.remotes {
			assert.Nil(test.Ping(c.netns, n))
		}
	}
}

func (m nodocker) checkNeighbor(t *testing.T) {
	assert := test.Assert{t}
	retries := 3
	var not_found bool
	time.Sleep(1 * time.Second)
	for i := retries; i > 0; i-- {
		not_found = false
		cmd := exec.Command("goes", "vnet", "show", "neigh")
		out, _ := cmd.Output()
		out_string := fmt.Sprintf("%s", out)
		for _, c := range m {
			for _, n := range c.remotes {
				re := regexp.MustCompile(n)
				match := re.FindAllStringSubmatch(out_string, -1)
				if len(match) == 0 {
					not_found = true
				}
			}
		}
		if not_found && *test.VV {
			fmt.Printf("%v", out_string)
			fmt.Printf("%v retries left\n", i-1)
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if not_found {
		test.Pause("Failed")
		assert.Nil(fmt.Errorf("no neighbor found"))
	}
}

// delete namespace without firt moving interface(s) out to default ns
// verify interface is now back in default namespace anyway
func (m nodocker) delNetns(t *testing.T) {
	assert := test.Assert{t}
	for _, c := range m {
		ns := c.netns
		_, err := os.Stat(filepath.Join("/var/run/netns", ns))
		if err == nil {
			assert.Program(test.Self{},
				"ip", "netns", "del", ns)
		}
	}
	time.Sleep(2 * time.Second)
	for k, _ := range m {
		ifname := netport.PortByNetPort[k]
		assert.Program(test.Self{},
			"ip", "link", "show", ifname)
	}
}

func (m nodocker) checkNoNeighbor(t *testing.T) {
	assert := test.Assert{t}
	cmd := exec.Command("goes", "vnet", "show", "neigh")
	out, _ := cmd.Output()
	out_string := fmt.Sprintf("%s", out)
	found := false
	for _, c := range m {
		for _, n := range c.remotes {
			re := regexp.MustCompile(n)
			match := re.FindAllStringSubmatch(out_string, -1)
			if len(match) > 0 {
				found = true
			}
		}
	}
	if found {
		if *test.VV {
			fmt.Printf("%v", out_string)
		}
		assert.Nil(fmt.Errorf("leftover neighbor found"))
	}
	// check leftover adjacencies as well
	err := test.NoAdjacency(t)
	assert.Nil(err)
}
