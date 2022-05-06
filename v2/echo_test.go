// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"strings"
	"testing"
)

func TestEcho(t *testing.T) {
	w := new(strings.Builder)
	ctx := WithOutput(context.Background(), w)
	try := func(t *testing.T, want string, args ...string) {
		t.Helper()
		w.Reset()
		if err := Echo(ctx, args...); err != nil {
			t.Error(err)
		} else if got := w.String(); got != want {
			t.Errorf("%q != %q", got, want)
		} else {
			t.Log(w.String())
		}
	}
	t.Run("with-trailing-newline", func(t *testing.T) {
		try(t, "hello world\n", "hello", "world")
	})
	t.Run("without-trailing-newline", func(t *testing.T) {
		try(t, "hello world", "-n", "hello", "world")
	})
	t.Run("with-leading-and-trailing-newline", func(t *testing.T) {
		try(t, "\nhello world\n", "-e", "\\nhello world")
	})
	t.Run("with-just-leading-newline", func(t *testing.T) {
		try(t, "\nhello world", "-e", "-n", "\\nhello world")
	})
}
