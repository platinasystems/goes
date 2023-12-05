// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"log"
	"net"
	"sync"
	"time"
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

	bx := New().Empty().ToAll().From(from).Close(bxc).Seal(bxc)
	defer bx.Recycle()

	t := time.NewTicker(period)
	defer t.Stop()

	for {
		if _, err := conn.Write(bx); err != nil {
			if err != context.Canceled {
				log.Print(err)
			}
			break
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
