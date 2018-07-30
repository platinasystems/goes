// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/platinasystems/go/internal/machine"
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/ethtool"
	"github.com/platinasystems/go/internal/test/netport"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/mk1"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/vnet"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/xeth"
)

var suite = test.Suite{
	Name: "goes-platina-mk1",
	Init: func(t *testing.T) {
		assert := test.Assert{t}
		assert.Main(main)
		assert.YoureRoot()
		assert.NoListener("@platina-mk1/vnetd")

		b, err := ioutil.ReadFile("/proc/net/unix")
		assert.Nil(err)
		if bytes.Index(b, []byte("@platina-mk1/xeth")) >= 0 {
			assert.Program("rmmod", "platina-mk1")
		}
		modprobe := []string{"modprobe", "platina-mk1"}
		const ko = "platina-mk1.ko"
		if _, err := os.Stat(ko); err == nil {
			modprobe = []string{"insmod", ko}
		}
		if *mk1.IsAlpha {
			modprobe = append(modprobe, "alpha=1")
		}
		if *test.VVV {
			modprobe = append(modprobe, "dyndbg=+pmf")
		} else {
			modprobe = append(modprobe, "dyndbg=-pmf")
		}
		assert.Program(modprobe)

		netport.Init(assert)
		ethtool.Init(assert)

		machine.Name = name
	},
	Tests: test.Tests{
		&xeth.Suite,
		&vnet.Suite,
	},
}

func Test(t *testing.T) {
	suite.Test(t)
}
