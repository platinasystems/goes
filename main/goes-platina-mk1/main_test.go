// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
	main "github.com/platinasystems/go/main/goes-platina-mk1"
)

func Test(t *testing.T) {
	if test.Goes {
		test.Exec(main.Goes().Main)
	}

	assert := test.Assert{t}
	assert.YoureRoot()
	assert.GoesNotRunning()

	defer assert.Background("goes", "redisd").Quit()
	assert.Program(12*time.Second, "goes", "hwait", "platina",
		"redis.ready", "true", "10")

	defer assert.Background(test.Debug{}, 30*time.Second,
		"goes", "vnetd").Quit()
	assert.Program(32*time.Second, "goes", "hwait", "platina",
		"vnet.ready", "true", "30")

	assert.Nil(docker.Check(t))

	test.Suite{
		{"ospf", test.Suite{
			{"eth", ospfEth},
			{"vlan", ospfVlan},
		}.Run},
		{"isis", test.Suite{
			{"eth", isisEth},
			{"vlan", isisVlan},
		}.Run},
		{"bgp", test.Suite{
			{"eth", bgpEth},
			{"vlan", bgpVlan},
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

func bgpEth(t *testing.T) {
	FrrBGP(t, "docs/examples/docker/frr-bgp/conf.yml")
}

func bgpVlan(t *testing.T) {
	FrrBGP(t, "docs/examples/docker/frr-bgp/conf_vlan.yml")
}

func netSlice(t *testing.T) {
	Slice(t, "docs/examples/docker/net-slice/conf_vlan.yml")
}
