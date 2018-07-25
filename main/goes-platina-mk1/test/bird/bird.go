// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bird

import (
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/bird/bgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/bird/ospf"
)

var Suite = test.Suite{
	Name: "bird",
	Tests: test.Tests{
		&bgp.Suite,
		&ospf.Suite,
	},
}
