// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package accept

import (
	"net"
	"sync"
)

// Send accepted connections to channel until listner is closed.
func Routine(wg *sync.WaitGroup, ch chan<- net.Conn, ln net.Listener) {
	defer wg.Done()
	defer close(ch)
	for {
		if c, err := ln.Accept(); err == nil {
			ch <- c
		} else {
			break
		}
	}
}
