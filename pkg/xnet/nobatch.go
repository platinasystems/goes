// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !android && !linux

package xnet

import (
	"context"
	"net"
)

const BatchCap = 1
const CanBatch = false

// This GOOS doesn't have “recvmmsg” so RecvBatchService calls
// [MsgPool.RecvService]
func (mp *MsgPool) RecvBatchService(
	ctx context.Context, ch chan<- *Msg, conn net.PacketConn,
) error {
	return mp.RecvService(ctx, ch, conn)
}

// This GOOS doesn't have “sendmmsg” so SendBatchService calls
// [MsgPool.SendService]
func (mp *MsgPool) SendBatchService(conn net.PacketConn, ch <-chan *Msg) error {
	return mp.SendService(conn, ch)
}
