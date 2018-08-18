// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnet

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/machine"
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/docker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/nodocker"
)

var redisd, vnetd *test.Program

var Suite = test.Suite{
	Name: "vnet",
	Init: func(t *testing.T) {
		assert := test.Assert{t}

		redisd = assert.Background(test.Self{}, "redisd")
		assert.Program(12*time.Second, test.Self{},
			"hwait", machine.Name, "redis.ready", "true", "10")

		vnetd = assert.Background(30*time.Second, test.Self{}, "vnetd")
		assert.Program(32*time.Second, test.Self{},
			"hwait", machine.Name, "vnet.ready", "true", "30")

		if *test.MustPause {
			test.Pause("Attach vnet debugger to pid(", vnetd.Pid(),
				");\nthen press enter to continue...")
		}
	},
	Exit: func(t *testing.T) {
		if redisd != nil {
			defer redisd.Quit()
		}
		if vnetd != nil {
			defer vnetd.Quit()
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
