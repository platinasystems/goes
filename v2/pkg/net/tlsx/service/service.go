// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	ErrIncomplete    = errors.New("incomplete")
	ErrMisconfigured = errors.New("misconfigured")
	ErrMissingInput  = errors.New("missing input")
)

var Select = func(
	context.Context,
	io.Reader,
	io.Writer,
	[]string,
	...string,
) error {
	return ErrMisconfigured
}

func Routine(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	addr := &net.TCPAddr{Port: port.RPC.ValueContext(ctx)}
	svr := rpc.NewServer()
	svr.Register(&Service{30 * time.Second})

	svch := make(chan *tls.Conn, 4)
	wg.Add(1)
	go greet.Routine(ctx, wg, svch, addr, &tls.Config{
		Certificates: []tls.Certificate{certs.Self.TLS()},
		ServerName:   certs.Self.Name(),
		ClientAuth:   tls.RequireAnyClientCert,
	})
	for c := range svch {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer c.Close()
			svr.ServeConn(c)
		}()
	}
}

type Service struct {
	timeout time.Duration
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
		return ErrIncomplete
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
	err := Select(ctx, r, w, path, args...)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	*result = w.String()
	return err
}
