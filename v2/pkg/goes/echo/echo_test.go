// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"context"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	ctx := context.Background()
	got := new(strings.Builder)
	path := []string{"echo.test", "echo"}
	try := func(t *testing.T, want string, args ...string) {
		t.Helper()
		got.Reset()
		err := Func(ctx, got, path, args...)
		if err != nil {
			t.Error(err)
		} else if gots := got.String(); gots != want {
			t.Errorf("%q != %q", gots, want)
		} else {
			t.Log(gots)
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
