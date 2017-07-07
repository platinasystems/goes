// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fe1

import (
	"github.com/platinasystems/go/vnet/ethernet"
)

// Platform configuration for FE1 based systems.
type Config struct {
	Version             uint
	BaseEthernetAddress ethernet.Address
	NEthernetAddress    uint
	Init                func()
}
