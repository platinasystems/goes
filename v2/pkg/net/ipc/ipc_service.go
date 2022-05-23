// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

// no handler timeout if 0
func (ipc Ipc) Service(
	ctx context.Context,
	wg *sync.WaitGroup,
	selector selection.Func,
	timeout time.Duration,
) error {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ln, err := ipc.Listen()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ln.Close()
		<-ctx.Done()
	}()

	svr := rpc.NewServer()
	svr.Register(&Service{
		ln.Addr().String(),
		selector,
		timeout,
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if c, err := ln.Accept(); err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					log.Println(err)
				}
			} else {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer c.Close()
					svr.ServeConn(c)
				}()
			}
		}
	}()

	return nil
}

type Service struct {
	address  string
	selector selection.Func
	timeout  time.Duration
}

func (svc *Service) Select(args []string, result *string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if svc.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, svc.timeout)
		defer cancel()
	}
	w := new(strings.Builder)
	var r io.Reader = io.LimitReader(nil, 0)
	if len(args) > 1 && args[0] == "-i" {
		if len(args) == 1 {
			return fmt.Errorf("missing input")
		}
		r = strings.NewReader(args[1])
		args = args[2:]
	}
	path := selection.Path{svc.address}
	err := svc.selector(ctx, r, w, path, args...)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	*result = w.String()
	return err
}
