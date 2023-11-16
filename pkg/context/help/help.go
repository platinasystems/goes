// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package help

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
	"github.com/platinasystems/goes/v2/pkg/flag"
)

var help parameter.Key[bool]

func Wanted(contextOrFlagSet ...any) bool {
	for _, v := range contextOrFlagSet {
		switch t := v.(type) {
		case context.Context:
			if help.Value(t) {
				return true
			}
		case *flag.FlagSet:
			if flag.Eval[bool](t, "help") {
				return true
			}
		}
	}
	return false
}

func With(ctx context.Context) context.Context {
	return help.With(ctx, true)
}
