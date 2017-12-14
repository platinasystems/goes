// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frr

import (
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/bgp"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/isis"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/frr/ospf"
)

var Suite = test.Suite{
	{"ospf", ospf.Suite},
	{"isis", isis.Suite},
	{"bgp", bgp.Suite},
}.Run
