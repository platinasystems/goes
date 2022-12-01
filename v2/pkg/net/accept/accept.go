// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package accept

import (
	"net"
	"sync"
)

// This is run as go routine to continually feed accepted connections to
// channel until the listner is closed.
func With(wg *sync.WaitGroup, ln net.Listener, ch chan<- net.Conn) {
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
