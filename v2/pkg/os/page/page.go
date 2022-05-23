// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package page

import (
	"os"
	"sync"
)

var (
	size = os.Getpagesize()
	pool = sync.Pool{
		New: func() any {
			return make([]byte, size, size)
		},
	}
)

func New() []byte {
	return pool.Get().([]byte)
}

func Free(pg []byte) {
	if cap(pg) == size {
		pg = pg[:size]
		pool.Put(pg)
	}
}
