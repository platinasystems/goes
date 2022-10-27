// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cat

import (
	"context"
	"strings"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

func Test(t *testing.T) {
	const want = "hello world\n"
	ctx := context.Background()
	got := new(strings.Builder)
	r := strings.NewReader(want)
	cat := selection.Path{"cat"}
	if err := Func(ctx, r, got, cat, "-"); err != nil {
		t.Error(err)
	} else if gots := got.String(); gots != want {
		t.Errorf("%q != %q", gots, want)
	} else {
		t.Log(gots)
	}
}
