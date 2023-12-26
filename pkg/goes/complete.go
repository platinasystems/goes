// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
)

var completeParameter bool

func ContextComplete(ctx context.Context) bool {
	return parameter.Value(ctx, &completeParameter)
}

func CompleteContext(ctx context.Context, t bool) context.Context {
	return parameter.Context(ctx, &completeParameter, t)
}
