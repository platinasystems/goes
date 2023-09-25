// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/log/style"
)

// Until cancelled, send periodic keep alives (empty boxes) to all members.
func KeepAlive(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.Conn,
	bxc *Cipher,
	from uint32,
	period time.Duration,
) {
	defer wg.Done()
	defer style.Recovery(context.Canceled)

	bx := New().Empty().ToAll().From(from).Close(bxc).Seal(bxc)
	defer bx.Recycle()

	t := time.NewTicker(period)
	defer t.Stop()

	for {
		if _, err := conn.Write(bx); err != nil {
			panic(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
