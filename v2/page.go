// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"os"
	"sync"
)

var (
	PageSz = os.Getpagesize()
	Page   = sync.Pool{
		New: func() interface{} {
			return make([]byte, PageSz, PageSz)
		},
	}
)
