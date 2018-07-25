// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net

import (
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net/dhcp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net/slice"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/net/static"
)

var Suite = test.Suite{
	Name: "net",
	Tests: test.Tests{
		&slice.Suite,
		&dhcp.Suite,
		&static.Suite,
	},
}
