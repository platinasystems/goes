// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package label

import "sync"

const (
	pageSize = 4 << 10
	pageMask = pageSize - 1
	byteBits = 8
	byteMask = byteBits - 1
)

type Book struct {
	sync.Mutex
	bits []pageOfLabels
}

type pageOfLabels []uint8

func newPageOfLabels() pageOfLabels {
	return make(pageOfLabels, pageSize, pageSize)
}

func (book *Book) New() Label {
	book.Lock()
	defer book.Unlock()

	for i, pg := range book.bits {
		for j, bits := range pg {
			for k := 0; k < byteBits; k++ {
				bit := uint8(1 << k)
				if (bits & bit) == 0 {
					pg[j] |= bit
					lbl := Label(i * pageSize)
					lbl += Label(j * byteBits)
					lbl += Label(k)
					return lbl
				}
			}
		}
	}

	pg := newPageOfLabels()
	pg[0] = 1
	lbl := Label(len(book.bits) * pageSize)
	book.bits = append(book.bits, pg)
	return lbl
}

func (book *Book) InUse(lbl Label) bool {
	book.Lock()
	defer book.Unlock()
	i, j, k := bookMark(lbl.Index())
	if i >= len(book.bits) || j >= pageSize {
		return false
	}
	return book.bits[i][j]&(1<<k) != 0
}

func (book *Book) Put(lbl Label) {
	book.Lock()
	defer book.Unlock()
	i, j, k := bookMark(lbl.Index())
	_ = book.bits[i][j]
	book.bits[i][j] &^= 1 << k
}

func (book *Book) Reserve(lbl Label) {
	book.Lock()
	defer book.Unlock()
	i, j, k := bookMark(lbl.Index())
	for pg := len(book.bits); pg <= i; pg++ {
		book.bits = append(book.bits, newPageOfLabels())
	}
	book.bits[i][j] |= 1 << k
}

// Page, Byte, Bit
func bookMark(lbl Index) (int, int, int) {
	return bookPage(lbl), pageByte(lbl), byteBit(lbl)
}

// Page index within label book.
func bookPage(lbl Index) int {
	return int(lbl) / pageSize
}

// Byte index within page of label book.
func pageByte(lbl Index) int {
	return (int(lbl) / byteBits) & pageMask
}

// Bit index within byte within page of label book.
func byteBit(lbl Index) int {
	return int(lbl) & byteMask
}
