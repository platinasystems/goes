// Copyright © 2018-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import "sync/atomic"

type Counter uint64

func (count *Counter) Count() uint64 {
	return atomic.LoadUint64((*uint64)(count))
}

func (count *Counter) Reset() {
	atomic.StoreUint64((*uint64)(count), 0)
}

func (count *Counter) Inc() {
	atomic.AddUint64((*uint64)(count), 1)
}
