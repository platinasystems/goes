// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

func TestIpc(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	sigctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancel(sigctx)
	defer cancel()

	r := io.LimitReader(nil, 0)
	got := new(strings.Builder)

	ipc := Preface("goes_ipc_test")
	path := selection.Path{ipc.String()}
	selector := selection.Map{
		"cat":  cat.Func,
		"echo": echo.Func,
	}.Select
	err := ipc.Service(ctx, &wg, selector, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ut := func(
		t *testing.T,
		want string,
		args ...string,
	) {
		t.Helper()
		got.Reset()
		err := ipc.Func(ctx, r, got, path, args...)
		if err != nil {
			t.Error(err)
		} else if gots := got.String(); gots != want {
			t.Errorf("%q != %q", gots, want)
		} else {
			t.Log(gots)
		}
	}
	t.Run("complete", func(t *testing.T) {
		ut(t, "echo\n", "complete", "ec")
	})
	t.Run("echo", func(t *testing.T) {
		ut(t, "hello world\n", "echo", "hello", "world")
	})
	t.Run("cat", func(t *testing.T) {
		const want = "hello world\n"
		r = strings.NewReader(want)
		ut(t, want, "-i", "-", "cat", "-")
	})
}
