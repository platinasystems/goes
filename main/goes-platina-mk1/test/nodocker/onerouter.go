// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nodocker

import (
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

// ping between 2 host namespaces throuh a router
func onerouter(t *testing.T) {
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}

	assert.Program(test.Self{}, "ip", "netns", "add", "h1")
	defer cleanup.Program(test.Self{}, "ip", "netns", "del", "h1")
	assert.Program(test.Self{}, "ip", "netns", "add", "r")
	defer cleanup.Program(test.Self{}, "ip", "netns", "del", "r")
	assert.Program(test.Self{}, "ip", "netns", "add", "h2")
	defer cleanup.Program(test.Self{}, "ip", "netns", "del", "h2")

	assert.Program(test.Self{}, "ip", "link", "set",
		netport.Map["net0port0"], "up", "netns", "h1")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "link", "set", netport.Map["net0port0"],
		"down", "netns", 1)
	assert.Program(test.Self{}, "ip", "link", "set",
		netport.Map["net0port1"], "up", "netns", "r")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "link", "set", netport.Map["net0port1"],
		"down", "netns", 1)
	assert.Program(test.Self{}, "ip", "link", "set",
		netport.Map["net1port0"], "up", "netns", "h2")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "link", "set", netport.Map["net1port0"],
		"down", "netns", 1)
	assert.Program(test.Self{}, "ip", "link", "set",
		netport.Map["net1port1"], "up", "netns", "r")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "link", "set", netport.Map["net1port1"],
		"down", "netns", 1)

	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "address", "add", "10.1.0.0/31",
		"dev", netport.Map["net0port0"])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "address", "del", "10.1.0.0/31",
		"dev", netport.Map["net0port0"])
	assert.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "address", "add", "10.1.0.1/31",
		"dev", netport.Map["net0port1"])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "address", "del", "10.1.0.1/31",
		"dev", netport.Map["net0port1"])
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "address", "add", "10.1.0.2/31",
		"dev", netport.Map["net1port0"])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "address", "del", "10.1.0.2/31",
		"dev", netport.Map["net1port0"])
	assert.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "address", "add", "10.1.0.3/31",
		"dev", netport.Map["net1port1"])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "r",
		test.Self{}, "ip", "address", "del", "10.1.0.3/31",
		"dev", netport.Map["net1port1"])

	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "route", "add", "10.1.0.2/31",
		"via", "10.1.0.1")
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "route", "add", "10.1.0.0/31",
		"via", "10.1.0.3")

	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ping", "10.1.0.1")
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ping", "10.1.0.3")
	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ping", "10.1.0.2")
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ping", "10.1.0.0")
}
