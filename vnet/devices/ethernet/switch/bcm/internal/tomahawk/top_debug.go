// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package tomahawk

import (
	. "github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/debug"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
)

func init() {
	// Check register addresses.
	r := (*top_regs)(m.RegsBasePointer)
	CheckRegAddr("core_pll0_control[0]", r.core_pll0_control[0].offset(), 0x38000)
	CheckRegAddr("temperature_sensor.control[0]", r.temperature_sensor.control[0].offset(), 0x50000)
	CheckRegAddr("core_pll_frequency_select", r.core_pll_frequency_select.offset(), 0x75c00)
}
