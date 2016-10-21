// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"
)

func GetHeader(r *vnet.Ref) *Header                 { return (*Header)(r.Data()) }
func GetPacketHeader(r *vnet.Ref) vnet.PacketHeader { return GetHeader(r) }
