// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hw

import (
	"github.com/platinasystems/go/elib"

	"unsafe"
)

var heap = &elib.MemHeap{}

func DmaInit(b []byte)                                            { heap.InitData(b) }
func DmaAlloc(n uint) (b []byte, id elib.Index, offset, cap uint) { return DmaAllocAligned(n, 0) }
func DmaFree(id elib.Index)                                       { heap.Put(id) }
func DmaGetData(id elib.Index) (b []byte)                         { return heap.GetId(id) }
func DmaGetPointer(o uint) unsafe.Pointer                         { return heap.Data(o) }
func DmaIsValidOffset(o uint) bool                                { return heap.OffsetValid(o) }
func DmaHeapUsage() string                                        { return heap.String() }
