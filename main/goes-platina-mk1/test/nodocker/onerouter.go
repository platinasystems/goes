// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nodocker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

// ping between 2 host namespaces throuh a router
func onerouter(t *testing.T) {
	type routeT struct {
		prefix string
		gw     string
	}
	ports := []struct {
		netns   string
		ifname  string
		ifa     string
		routes  []routeT
		remotes []string
	}{
		{
			netns:  "h1",
			ifname: netport.Map["net0port0"],
			ifa:    "10.1.0.0/31",
			routes: []routeT{
				{"10.1.0.2/31", "10.1.0.1"},
			},
			remotes: []string{"10.1.0.2"},
		},
		{
			netns:  "r",
			ifname: netport.Map["net0port1"],
			ifa:    "10.1.0.1/31",
		},
		{
			netns:  "h2",
			ifname: netport.Map["net1port0"],
			ifa:    "10.1.0.2/31",
			routes: []routeT{
				{"10.1.0.0/31", "10.1.0.3"},
			},
			remotes: []string{"10.1.0.0"},
		},
		{
			netns:  "r",
			ifname: netport.Map["net1port1"],
			ifa:    "10.1.0.3/31",
		},
	}

	assert := test.Assert{t}
	cleanup := test.Cleanup{t}

	for _, port := range ports {
		_, err := os.Stat(filepath.Join("/var/run/netns", port.netns))
		if err != nil {
			assert.Program(test.Self{},
				"ip", "netns", "add", port.netns)
			defer cleanup.Program(test.Self{},
				"ip", "netns", "del", port.netns)
		}
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
		for _, route := range port.routes {
			assert.Program(test.Self{},
				"ip", "netns", "exec", port.netns, test.Self{},
				"ip", "route", "add", route.prefix,
				"via", route.gw)
		}
	}
	for _, port := range ports {
		assert.Nil(test.Carrier(port.netns, port.ifname))
	}
	for _, port := range ports {
		for _, route := range port.routes {
			assert.Nil(test.Ping(port.netns, route.gw))
		}
	}
	for _, port := range ports {
		for _, remote := range port.remotes {
			assert.Nil(test.Ping(port.netns, remote))
		}
	}
}
