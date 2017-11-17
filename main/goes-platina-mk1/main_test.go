// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
	main "github.com/platinasystems/go/main/goes-platina-mk1"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/bgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/isis"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/ospf"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net/slice"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/port2port"
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

func bgpEth(t *testing.T) {
	bgp.Test(t, conf(t, "bgp", bgp.Conf))
}

func bgpVlan(t *testing.T) {
	bgp.Test(t, conf(t, "bgp-vlan", bgp.ConfVlan))
}

func isisEth(t *testing.T) {
	isis.Test(t, conf(t, "isis", isis.Conf))
}

func isisVlan(t *testing.T) {
	isis.Test(t, conf(t, "isis-vlan", isis.ConfVlan))
}

func netSlice(t *testing.T) {
	slice.Test(t, conf(t, "net-slice", slice.Conf))
}

func ospfEth(t *testing.T) {
	ospf.Test(t, conf(t, "ospf", ospf.Conf))
}

func ospfVlan(t *testing.T) {
	ospf.Test(t, conf(t, "ospf-vlan", ospf.ConfVlan))
}

func conf(t *testing.T, name, text string) []byte {
	assert := test.Assert{t}
	assert.Helper()
	tmpl, err := template.New(name).Parse(text)
	assert.Nil(err)
	buf := new(bytes.Buffer)
	assert.Nil(tmpl.Execute(buf, port2port.Conf))
	return buf.Bytes()
}
