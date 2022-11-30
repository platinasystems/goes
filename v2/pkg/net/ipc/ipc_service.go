// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"errors"
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
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var ErrMissingInput = errors.New("missing input")

// no handler timeout if 0
func (ipc Ipc) Service(
	ctx context.Context,
	wg *sync.WaitGroup,
	m selection.Map,
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
	svr.Register(&Service{ln.Addr().String(), timeout, m})

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
	address string
	timeout time.Duration
	m       selection.Map
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
	path := []string{program.Base()}
	if len(args) == 0 {
		return selection.ErrIncomplete
	}
	if args[0] == "complete" {
		path = append(path, args[0])
		if args = args[1:]; len(args) > 1 && args[0] == "help" {
			args = args[1:]
		}
	} else if args[0] == "help" {
		path = append(path, args[0])
		args = args[1:]
	} else if args[0] == "-i" {
		if len(args) == 1 {
			return ErrMissingInput
		}
		r = strings.NewReader(args[1])
		args = args[2:]
	}
	path = append(path, svc.address)
	err := svc.m.Select(ctx, r, w, path, args...)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	*result = w.String()
	return err
}
