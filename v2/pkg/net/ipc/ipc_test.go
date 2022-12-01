// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/service"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/foreclose"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

func TestIpc(t *testing.T) {
	const timeout = 30 * time.Second

	var wg sync.WaitGroup
	defer wg.Wait()

	sigctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancel(sigctx)
	defer cancel()

	r := io.LimitReader(nil, 0)
	got := new(strings.Builder)

	ipc := New()
	req := service.Request{ipc}
	ln, err := ipc.Listen()
	if err != nil {
		t.Fatal(err)
	}
	lna := ln.Addr().String()

	wg.Add(1)
	go foreclose.With(ctx, &wg, ln)

	conch := make(chan net.Conn, 4)

	wg.Add(1)
	go accept.With(&wg, ln, conch)

	wg.Add(1)
	go service.With(&wg, conch, lna, timeout, selection.Map{
		"cat":  cat.Func,
		"echo": echo.Func,
	})

	prog := program.Base()
	ipcs := ipc.String()
	ut := func(
		t *testing.T,
		want string,
		path []string,
		args ...string,
	) {
		t.Helper()
		got.Reset()
		err := req.Func(ctx, r, got, path, args...)
		if err != nil {
			t.Error(err)
		} else if gots := got.String(); gots != want {
			t.Errorf("%q != %q", gots, want)
		} else {
			t.Log(gots)
		}
	}
	t.Run("complete", func(t *testing.T) {
		ut(t, "echo\n", []string{prog, "complete", ipcs}, "ec")
	})
	t.Run("echo", func(t *testing.T) {
		ut(t, "hello world\n", []string{prog, ipcs},
			"echo", "hello", "world")
	})
	t.Run("cat", func(t *testing.T) {
		const want = "hello world\n"
		r = strings.NewReader(want)
		ut(t, want, []string{prog, ipcs}, "-i", "-", "cat", "-")
	})
}
