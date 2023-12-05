// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package pathctx

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var Parameter = context.NewParameter([]string{program.Base()})

func AppendIn(ctx context.Context, names ...string) context.Context {
	if len(names) > 0 {
		ctx = Parameter.With(ctx, append(Parameter.In(ctx), names...))
	}
	return ctx
}

func StringIn(ctx context.Context) string {
	return strings.Join(Parameter.In(ctx), " ")
}
