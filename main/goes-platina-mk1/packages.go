// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/plugins/fe1"
)

func init() { Packages = fe1.Packages }
