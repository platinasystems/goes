// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package page

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/sync/chunk"
	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

var (
	size  = os.Getpagesize()
	Align = align.Align(size).Roundup
	Free  = chunk.Free
	New   = func() []byte { return chunk.New(size) }
)

func Size() int { return size }

func init() {
	switch size {
	case chunk.Size4K:
		Free = chunk.Free4K
		New = func() []byte { return chunk.New4K() }
	case chunk.Size8K:
		Free = chunk.Free8K
		New = func() []byte { return chunk.New8K() }
	}
}
