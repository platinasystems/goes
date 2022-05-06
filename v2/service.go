// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const ServiceTimeout = 30 * time.Second

func Listen() (net.Listener, error) { return net.Listen(RpcNet, RpcAddr) }

func (m Selection) Service(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
) {
	defer wg.Done()
	svr := rpc.NewServer()
	svr.Register(&Service{ln.Addr().String(), m})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				svr.ServeConn(conn)
			}()
		}
	}()
	<-ctx.Done()
	ln.Close()
}

type Service struct {
	lns string
	m   Selection
}

func (svc *Service) Select(
	args []string,
	result *string,
) error {
	w := new(strings.Builder)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, ServiceTimeout)
	defer cancel()
	ctx = WithRoot(ctx, svc.m)
	ctx = WithOutput(ctx, w)
	ctx = WithPath(ctx, svc.lns)
	ctx, args = Hereby(ctx, args)
	ctx, args = Preempt(ctx, args)
	err := svc.m.Select(ctx, args...)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	*result = w.String()
	return err
}
