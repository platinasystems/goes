// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package foreclose

import (
	"context"
	"io"
	"sync"
)

// Close when context is done.
func Routine(ctx context.Context, wg *sync.WaitGroup, v io.Closer) {
	defer wg.Done()
	defer v.Close()
	<-ctx.Done()
}
