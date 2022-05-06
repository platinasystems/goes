// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
)

func TestRPC(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	w := new(strings.Builder)
	ctx = WithOutput(ctx, w)
	ctx = WithPath(ctx, Prog)
	ln, err := Listen()
	if err != nil {
		t.Fatal(err)
	}
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg.Add(1)
	go BuiltIn.Service(cctx, &wg, ln)
	ut := func(
		t *testing.T,
		want string,
		args ...string,
	) {
		t.Helper()
		w.Reset()
		if err := RPC(ctx, args...); err != nil {
			t.Error(err)
		} else if got := w.String(); got != want {
			t.Errorf("%q != %q", got, want)
		}
	}
	t.Run("complete", func(t *testing.T) {
		ut(t, "echo\n", "complete", "ec")
	})
	t.Run("echo", func(t *testing.T) {
		ut(t, "hello world\n", "echo", "hello", "world")
	})
	t.Run("cat", func(t *testing.T) {
		ctx = WithInput(ctx, strings.NewReader("hello world\n"))
		ut(t, "hello world\n", "cat", "-", "<")
	})
}
