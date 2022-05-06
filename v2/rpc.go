// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"net"
	"net/rpc"
)

func RPC(ctx context.Context, args ...string) error {
	var res string
	args = append(PathOf(ctx)[1:], args...)
	ctx, args = Herein(ctx, args)
	conn, err := net.Dial(RpcNet, RpcAddr)
	if err != nil {
		return err
	}
	c := rpc.NewClient(conn)
	call := c.Go("Service.Select", args, &res, nil)
	select {
	case <-call.Done:
		err = call.Error
		c.Close()
	case <-ctx.Done():
		err = ctx.Err()
		c.Close()
		<-call.Done
	}
	if len(res) > 0 {
		OutputOf(ctx).Print(res)
	}
	return err
}
