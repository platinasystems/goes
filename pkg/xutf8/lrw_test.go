// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xutf8

import (
	"fmt"
	"strings"
	"testing"
)

func TestLastRuneWapper(t *testing.T) {
	sb := new(strings.Builder)
	lrw := NewLastRuneWrapper(sb)
	fmt.Fprint(lrw, "hello")
	if lrw.LastWrittenRune() == '\n' {
		t.Fatalf("unexpected newline: %q", sb.String())
	}
	fmt.Fprintln(lrw, " world")
	if lrw.LastWrittenRune() != '\n' {
		t.Fatalf("expected newline: %q", sb.String())
	}
	fmt.Fprint(lrw, "bon jour")
	if lrw.LastWrittenRune() == '\n' {
		t.Fatalf("unexpected newline: %q", sb.String())
	}
}
