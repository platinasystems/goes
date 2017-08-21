// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"unsafe"
)

//go:noescape
func getPC(argp unsafe.Pointer, PCHashSeed uint64) (Time, PC, PCHash uint64)
