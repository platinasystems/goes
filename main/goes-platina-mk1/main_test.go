// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"flag"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	main "github.com/platinasystems/go/main/goes-platina-mk1"
)

var loopback bool

func init() {
	flag.BoolVar(&loopback, "test.loopback", false,
		"run goes loopback tests")
}

func Test(t *testing.T) {
	if test.Goes {
		test.Exec(main.Goes().Main)
	}

	assert := test.Assert{t}
	assert.YoureRoot()
	if !loopback {
		t.Skip("need -test.loopback yaml conf file")
	}

	err := test.CheckDocker(t)
	if err != nil {
		t.Skip("no docker")
	}

	defer assert.Program(nil,
		"goes", "redisd",
	).Quit(10 * time.Second)

	assert.Program(nil,
		"goes", "hwait", "platina", "redis.ready", "true", "10",
	).Ok()

	defer assert.Program(nil,
		"goes", "vnetd",
	).Gdb().Quit(30 * time.Second)

	assert.Program(nil,
		"goes", "hwait", "platina", "vnet.ready", "true", "30",
	).Ok()

	test.Suite{
		{"ospf", test.Suite{
			{"eth", ospfEth},
			{"vlan", ospfVlan},
		}.Run},
		{"isis", test.Suite{
			{"eth", isisEth},
			{"vlan", isisVlan},
		}.Run},
		{"net-slice", netSlice},
	}.Run(t)
}

func ospfEth(t *testing.T) {
	FrrOSPF(t, "docs/examples/docker/frr-ospf/conf.yml")
}

func ospfVlan(t *testing.T) {
	FrrOSPF(t, "docs/examples/docker/frr-ospf/conf_vlan.yml")
}

func isisEth(t *testing.T) {
	FrrISIS(t, "docs/examples/docker/frr-isis/conf.yml")
}

func isisVlan(t *testing.T) {
	FrrISIS(t, "docs/examples/docker/frr-isis/conf_vlan.yml")
}

func netSlice(t *testing.T) {
	Slice(t, "docs/examples/docker/net-slice/conf_vlan.yml")
}
