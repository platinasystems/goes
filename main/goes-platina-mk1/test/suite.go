// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/ethtool"
	"github.com/platinasystems/go/internal/test/netport"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/docker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/mk1"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/nodocker"
	"github.com/platinasystems/redis"
)

const TestDir = "github.com/platinasystems/go/main/goes-platina-mk1"

var redisdProgram, vnetdProgram *test.Program

func Suite(Machine string, main func(), t *testing.T) {
	test.Suite{
		Name: Machine,
		Init: func(t *testing.T) {
			assert := test.Assert{t}
			assert.Dir(TestDir)
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

			redis.DefaultHash = Machine

			redisdProgram = assert.Background(test.Self{},
				"redisd")
			assert.Program(12*time.Second, test.Self{},
				"hwait", redis.DefaultHash, "redis.ready", "true", "10")

			vnetdProgram = assert.Background(30*time.Second, test.Self{},
				"vnetd")
			assert.Program(32*time.Second, test.Self{},
				"hwait", redis.DefaultHash, "vnet.ready", "true", "30")

			test.Pause("attach vnet debugger to pid ", vnetdProgram.Pid())
		},
		Exit: func(t *testing.T) {
			if redisdProgram != nil {
				defer redisdProgram.Quit()
			}
			if vnetdProgram != nil {
				defer vnetdProgram.Quit()
			}
			test.Pause("tests complete")
		},
		Tests: test.Tests{
			&test.Unit{"ready", func(*testing.T) {}},
			&nodocker.Suite,
			&docker.Suite,
		},
	}.Test(t)
}
