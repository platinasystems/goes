// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"context"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

// This returns a selection.Func that prints it's receiver.
func Func(v any) selection.Func {
	return show{v}.funk
}

type show struct {
	v any
}

func (sh show) funk(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "\nPrint named value.")
		return nil
	}
	fmt.Fprintln(w, sh.v)
	return ctx.Err()
}
