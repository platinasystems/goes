// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package accept

import (
	"context"
	"net"
)

// This starts a pair of go-routines that forward accepted connections until
// ctx.Done() whence they close both listener and forwarding channel. Usage,
//
//	for c := range accept.With(ctx, ln, make(chan net.Conn, depth) { ... }
func With(
	ctx context.Context,
	ln net.Listener,
	ch chan net.Conn,
) <-chan net.Conn {
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	go func() {
		for c, err := ln.Accept(); err == nil; c, err = ln.Accept() {
			ch <- c
		}
		close(ch)
	}()
	return ch
}
