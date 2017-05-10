// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes/cmd/license"
	"github.com/platinasystems/go/goes/cmd/patents"
	"github.com/platinasystems/go/goes/cmd/version"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/plugin/fe1"
)

func init() {
	license.Packages = fe1.Packages
	patents.Packages = fe1.Packages
	version.Packages = fe1.Packages
}
