// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
        "github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/lceventsd"
)


func init() {
	lceventsd.Init = func() {
                // lcabsd's io signals via pca9539 io-expander
                lceventsd.VdevIo = lceventsd.I2cDev{0, 0x76, 0, 0x72, 0x01, 0, 0, 0}
	}
}
