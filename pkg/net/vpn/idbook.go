// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
)

const (
	IdBookPageSize = 4 << 10
	IdBookPageMask = IdBookPageSize - 1
	IdBookByteBits = 8
	IdBookByteMask = IdBookByteBits - 1
)

type IdBook struct {
	sync.Mutex
	bits []IdBookPage
}

type IdBookPage []uint8

func NewIdBookPage() IdBookPage {
	return make(IdBookPage, IdBookPageSize, IdBookPageSize)
}

func (book *IdBook) New() box.Id {
	book.Lock()
	defer book.Unlock()

	for i, pg := range book.bits {
		for j, bits := range pg {
			for k := 0; k < IdBookByteBits; k++ {
				bit := uint8(1 << k)
				if (bits & bit) == 0 {
					pg[j] |= bit
					id := box.Id(i * IdBookPageSize)
					id += box.Id(j * IdBookByteBits)
					id += box.Id(k)
					return id
				}
			}
		}
	}

	pg := NewIdBookPage()
	pg[0] = 1
	id := box.Id(len(book.bits) * IdBookPageSize)
	book.bits = append(book.bits, pg)
	return id
}

func (book *IdBook) InUse(id box.Id) bool {
	book.Lock()
	defer book.Unlock()
	i, j, k := IdBookMark(IdIndex(id))
	if i >= len(book.bits) || j >= IdBookPageSize {
		return false
	}
	return book.bits[i][j]&(1<<k) != 0
}

func (book *IdBook) Put(id box.Id) {
	book.Lock()
	defer book.Unlock()
	i, j, k := IdBookMark(IdIndex(id))
	_ = book.bits[i][j]
	book.bits[i][j] &^= 1 << k
}

func (book *IdBook) Reserve(id box.Id) {
	book.Lock()
	defer book.Unlock()
	i, j, k := IdBookMark(IdIndex(id))
	for pg := len(book.bits); pg <= i; pg++ {
		book.bits = append(book.bits, NewIdBookPage())
	}
	book.bits[i][j] |= 1 << k
}

// Page, Byte, Bit
func IdBookMark(i int) (int, int, int) {
	return IdBookPageIndex(i), IdBookByteIndex(i), IdBookBitIndex(i)
}

// Page index within book.
func IdBookPageIndex(idi int) int {
	return idi / IdBookPageSize
}

// Byte index within page of book.
func IdBookByteIndex(idi int) int {
	return (idi / IdBookByteBits) & IdBookPageMask
}

// Bit index within byte within page of book.
func IdBookBitIndex(idi int) int {
	return idi & IdBookByteMask
}
