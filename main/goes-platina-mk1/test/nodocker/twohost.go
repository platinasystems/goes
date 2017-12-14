// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nodocker

import (
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/port2port"
)

// ping between 2 host namespaces
func twohost(t *testing.T) {
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}

	assert.Program(test.Self{}, "ip", "netns", "add", "h1")
	defer cleanup.Program(test.Self{}, "ip", "netns", "del", "h1")
	assert.Program(test.Self{}, "ip", "netns", "add", "h2")
	defer cleanup.Program(test.Self{}, "ip", "netns", "del", "h2")

	assert.Program(test.Self{}, "ip", "link", "set",
		port2port.Conf[0][0], "netns", "h1", "up")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "link", "set", port2port.Conf[0][0], "down")
	assert.Program(test.Self{}, "ip", "link", "set",
		port2port.Conf[0][1], "netns", "h2", "up")
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "link", "set", port2port.Conf[0][1], "down")

	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "address", "add", "10.1.0.0/31",
		"dev", port2port.Conf[0][0])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h1",
		test.Self{}, "ip", "address", "del", "10.1.0.0/31",
		"dev", port2port.Conf[0][0])
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "address", "add", "10.1.0.1/31",
		"dev", port2port.Conf[0][1])
	defer cleanup.Program(test.Self{}, "ip", "netns", "exec", "h2",
		test.Self{}, "ip", "address", "del", "10.1.0.1/31",
		"dev", port2port.Conf[0][1])

	assert.Program(test.Self{}, "ip", "netns", "exec", "h1",
		"ping", "-c", 3, "10.1.0.1")
	assert.Program(test.Self{}, "ip", "netns", "exec", "h2",
		"ping", "-c", 3, "10.1.0.0")
}
