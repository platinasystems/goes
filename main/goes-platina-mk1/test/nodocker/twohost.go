// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nodocker

import (
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

// ping between 2 host namespaces
func twohost(t *testing.T) {
	ports := []struct {
		netns  string
		ifname string
		ifa    string
		peer   string
	}{
		{
			netns:  "h1",
			ifname: netport.PortByNetPort["net0port0"],
			ifa:    "10.1.0.0/31",
			peer:   "10.1.0.1",
		},
		{
			netns:  "h2",
			ifname: netport.PortByNetPort["net0port1"],
			ifa:    "10.1.0.1/31",
			peer:   "10.1.0.0",
		},
	}

	assert := test.Assert{t}
	cleanup := test.Cleanup{t}

	for _, port := range ports {
		assert.Program(test.Self{},
			"ip", "netns", "add", port.netns)
		defer cleanup.Program(test.Self{},
			"ip", "netns", "del", port.netns)
		assert.Program(test.Self{},
			"ip", "link", "set", port.ifname, "up",
			"netns", port.netns)
		defer cleanup.Program(test.Self{},
			"ip", "netns", "exec", port.netns, test.Self{},
			"ip", "link", "set", port.ifname, "down", "netns", 1)
		assert.Program(test.Self{},
			"ip", "netns", "exec", port.netns, test.Self{},
			"ip", "address", "add", port.ifa, "dev", port.ifname)
		defer cleanup.Program(test.Self{},
			"ip", "netns", "exec", port.netns, test.Self{},
			"ip", "address", "del", port.ifa, "dev", port.ifname)
	}
	for _, port := range ports {
		assert.Nil(test.Carrier(port.netns, port.ifname))
	}
	for _, port := range ports {
		assert.Nil(test.Ping(port.netns, port.peer))
	}
}
