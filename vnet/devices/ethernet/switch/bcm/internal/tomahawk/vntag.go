// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
)

type vntag_action uint8

const (
	vntag_action_none vntag_action = iota
	vntag_action_change
	vntag_action_change_etag
	vntag_action_delete
)

func (a *vntag_action) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(a), b, i+1, i, isSet)
}
