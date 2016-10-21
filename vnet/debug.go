// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package vnet

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"unsafe"
)

func init() {
	if got, want := unsafe.Sizeof(Ref{}), unsafe.Sizeof(hw.Ref{}); got != want {
		panic(fmt.Errorf("ref size %d %d", got, want))
	}
}
