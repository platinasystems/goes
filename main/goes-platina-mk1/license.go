// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/internal/goes/cmd/license"
	"github.com/platinasystems/go/internal/goes/cmd/patents"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/copyright"
)

func init() {
	const dir = "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	license.Others = []license.Other{{dir, copyright.License}}
	patents.Others = []patents.Other{{dir, copyright.Patents}}
}
