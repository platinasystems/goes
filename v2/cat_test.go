// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"strings"
	"testing"
)

func TestCat(t *testing.T) {
	const want = "hello world\n"
	w := new(strings.Builder)
	ctx := WithInput(context.Background(), strings.NewReader(want))
	ctx = WithOutput(ctx, w)
	if err := Cat(ctx, "-"); err != nil {
		t.Error(err)
	} else if got := w.String(); got != want {
		t.Errorf("%q != %q", got, want)
	} else {
		t.Log(w.String())
	}
}
