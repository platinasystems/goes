// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package debug

import (
	"fmt"
)

func CheckAddr(name string, got, want uint) {
	if got != want {
		panic(fmt.Errorf("%s got 0x%x != want 0x%x", name, got, want))
	}
}
