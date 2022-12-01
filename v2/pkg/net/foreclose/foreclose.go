// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package foreclose

import (
	"context"
	"net"
	"sync"
)

// Close listener when context is done.
func With(ctx context.Context, wg *sync.WaitGroup, ln net.Listener) {
	defer wg.Done()
	defer ln.Close()
	<-ctx.Done()
}
