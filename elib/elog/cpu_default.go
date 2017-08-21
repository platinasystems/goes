// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64

package elog

import (
	"github.com/platinasystems/go/elib/cpu"

	"time"
)

func getCallerPC(argp unsafe.Pointer) uint64

func rotl_31(x uint64) uint64 { return (x << 31) | (x >> (64 - 31)) }
func pcHash(pc, seed uint64) uint {
	const (
		// Constants for multiplication: four random odd 64-bit numbers.
		m1 = 16877499708836156737
		m2 = 2820277070424839065
		m3 = 9497967016996688599
	)
	h := pc ^ seed
	h = rotl_31(h*m1) * m2
	h ^= h >> 29
	h *= m3
	h ^= h >> 32
	return uint(h)
}

func getPC(argp unsafe.Pointer, pcHashSeed uint64) (time, pc, pcHash uint64) {
	time = uint64(cpu.TimeNow())
	pc = getCallerPC(argp)
	pcHash = pcHash(pc, pcHashSeed)
	return
}
