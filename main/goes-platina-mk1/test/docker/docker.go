// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package docker

import (
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/docker"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/bird"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/gobgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net"
)

var Suite = test.Suite{
	Name: "docker",
	Init: func(t *testing.T) {
		err := docker.Check(t)
		if err != nil {
			t.Skip(err)
		}
	},
	Tests: test.Tests{
		&bird.Suite,
		&frr.Suite,
		&gobgp.Suite,
		&net.Suite,
	},
}
