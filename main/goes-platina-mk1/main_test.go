// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/bgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/isis"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/ospf"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net/slice"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/port2port"
)

func Test(t *testing.T) {
	test.Main(main)

	assert := test.Assert{t}
	assert.YoureRoot()
	assert.GoesNotRunning()

	defer assert.Background(test.Self{}, "redisd").Quit()
	assert.Program(12*time.Second, test.Self{}, "hwait", "platina",
		"redis.ready", "true", "10")

	defer assert.Background(30*time.Second, test.Self{}, test.Debug{},
		"vnetd").Quit()
	assert.Program(32*time.Second, test.Self{}, "hwait", "platina",
		"vnet.ready", "true", "30")

	assert.Nil(docker.Check(t))

	test.Suite{
		{"ospf", test.Suite{
			{"eth", func(t *testing.T) {
				ospf.Test(t, conf(t, "ospf", ospf.Conf))
			}},
			{"vlan", func(t *testing.T) {
				ospf.Test(t, conf(t, "ospf-vlan",
					ospf.ConfVlan))
			}},
		}.Run},
		{"isis", test.Suite{
			{"eth", func(t *testing.T) {
				isis.Test(t, conf(t, "isis", isis.Conf))
			}},
			{"vlan", func(t *testing.T) {
				isis.Test(t, conf(t, "isis-vlan",
					isis.ConfVlan))
			}},
		}.Run},
		{"bgp", test.Suite{
			{"eth", func(t *testing.T) {
				bgp.Test(t, conf(t, "bgp", bgp.Conf))
			}},
			{"vlan", func(t *testing.T) {
				bgp.Test(t, conf(t, "bgp-vlan", bgp.ConfVlan))
			}},
		}.Run},
		{"net-slice", test.Suite{
			{"vlan", func(t *testing.T) {
				slice.Test(t, conf(t, "net-slice-vlan",
					slice.ConfVlan))
			}},
		}.Run},
	}.Run(t)
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
