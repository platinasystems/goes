// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/vnet/devices/optics/sfp"

	"fmt"
)

func main() {
	m := &sfp.QsfpModule{
		BusIndex:   10,
		BusAddress: 20,
	}
	m.Present()
	fmt.Printf("%+v\n", m)
}
