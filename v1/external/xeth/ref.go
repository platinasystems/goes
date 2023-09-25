// Copyright © 2018-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import "sync/atomic"

type Ref int32

func (ref *Ref) Hold() int32 {
	return atomic.AddInt32((*int32)(ref), 1)
}

func (ref *Ref) Release() int32 {
	return atomic.AddInt32((*int32)(ref), -1)
}

func (ref *Ref) Count() int32 {
	return atomic.LoadInt32((*int32)(ref))
}

func (ref *Ref) Reset() {
	atomic.StoreInt32((*int32)(ref), 0)
}
