// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "context"

type Value struct {
	V interface{}
}

func (val Value) Show(ctx context.Context, _ ...string) error {
	switch Preemption(ctx) {
	case "":
	case "help":
		Usage(ctx, "\nPrint named value.")
		fallthrough
	default:
		return nil
	}
	OutputOf(ctx).Println(val.V)
	return ctx.Err()
}
