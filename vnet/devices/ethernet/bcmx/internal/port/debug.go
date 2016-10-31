// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package port

import (
	. "github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/debug"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
)

// Check memory map.
func init() {
	r := (*clport_regs)(m.RegsBasePointer)
	CheckRegAddr("tsc_uc_data_access_mode", r.tsc_uc_data_access_mode.offset(), 0x21900)
}
