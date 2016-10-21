// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"runtime"
)

// Name of current function as string.
func FuncName() (n string) {
	if pc, _, _, ok := runtime.Caller(1); ok {
		n = runtime.FuncForPC(pc).Name()
	} else {
		n = "unknown"
	}
	return
}
