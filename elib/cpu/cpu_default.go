// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64

package cpu

import (
	"time"
)

// Cache lines on generic.
const Log2CacheLineBytes = 6

func TimeNow() Time {
	return Time(time.Now().UnixNano())
}
func GetCallerPC() uintptr {
	panic("not implemented")
}
