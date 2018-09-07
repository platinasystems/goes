// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/machine"
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/ethtool"
	"github.com/platinasystems/go/internal/test/netport"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/docker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/mk1"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/nodocker"
)

var redisdProgram, vnetdProgram *test.Program

var suite = test.Suite{
	Name: "goes-platina-mk1",
	Init: func(t *testing.T) {
		assert := test.Assert{t}
		assert.Dir("github.com/platinasystems/go/main/goes-platina-mk1")
		assert.Main(main)
		assert.YoureRoot()
		assert.NoListener("@platina-mk1/vnetd")

		b, err := ioutil.ReadFile("/proc/net/unix")
		assert.Nil(err)
		if bytes.Index(b, []byte("@xeth")) >= 0 {
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

		redisdProgram = assert.Background(test.Self{},
			"redisd")
		assert.Program(12*time.Second, test.Self{},
			"hwait", machine.Name, "redis.ready", "true", "10")

		vnetdProgram = assert.Background(30*time.Second, test.Self{},
			"vnetd")
		assert.Program(32*time.Second, test.Self{},
			"hwait", machine.Name, "vnet.ready", "true", "30")

		if *test.MustPause {
			test.Pause("Attach vnet debugger to pid(",
				vnetdProgram.Pid(),
				");\nthen press enter to continue...")
		}
	},
	Exit: func(t *testing.T) {
		if redisdProgram != nil {
			defer redisdProgram.Quit()
		}
		if vnetdProgram != nil {
			defer vnetdProgram.Quit()
		}
		if *test.MustPause {
			test.Pause("press enter to continue...")
		}
	},
	Tests: test.Tests{
		&test.Unit{"ready", func(*testing.T) {}},
		&nodocker.Suite,
		&docker.Suite,
	},
}

func Test(t *testing.T) {
	suite.Test(t)
}
