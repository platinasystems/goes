// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
	"github.com/platinasystems/go/internal/test/ethtool"
	"github.com/platinasystems/go/internal/test/netport"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/bird"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/gobgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/nodocker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/xeth"
)

var testPause = flag.Bool("test.pause", false, "pause before and after suite")
var testCD = flag.String("test.cd", ".",
	"change to named directory before running tests")

func TestXeth(t *testing.T) {
	test.Main(main)

	assert := test.Assert{t}
	assert.YoureRoot()
	assert.NoListener("@platina-mk1/vnetd")
	t.Run("xeth", xeth.TestNames)
}

func Test(t *testing.T) {
	if *testCD != "." {
		if err := os.Chdir(*testCD); err != nil {
			t.Fatal(err)
		}
	}

	test.Main(main)

	assert := test.Assert{t}
	assert.YoureRoot()
	assert.NoListener("@platina-mk1/vnetd")
	netport.Init(assert)
	ethtool.Init(assert)

	defer assert.Background(test.Self{}, "redisd").Quit()
	assert.Program(12*time.Second, test.Self{}, "hwait", name,
		"redis.ready", "true", "10")

	vnetd := assert.Background(30*time.Second, test.Self{}, "vnetd")
	defer vnetd.Quit()
	assert.Program(32*time.Second, test.Self{}, "hwait", name,
		"vnet.ready", "true", "30")

	if *testPause {
		test.Pause("Attach vnet debugger to pid(", vnetd.Pid(), ");\n",
			"then press enter to continue...")
		defer test.Pause("complete, press enter to continue...")
	}

	test.Suite{
		{"vnet.ready", func(*testing.T) {}},
		{"nodocker", nodocker.Suite},
		{"docker", func(t *testing.T) {
			err := docker.Check(t)
			if err != nil {
				t.Skip(err)
			}
			test.Suite{
				{"net", net.Suite},
				{"frr", frr.Suite},
				{"gobgp", gobgp.Suite},
				{"bird", bird.Suite},
			}.Run(t)
		}},
	}.Run(t)
}
