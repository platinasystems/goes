// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
)

var helpParameter bool

func ContextHelp(ctx context.Context) bool {
	return parameter.Value(ctx, &helpParameter)
}

func HelpContext(ctx context.Context, t bool) context.Context {
	return parameter.Context(ctx, &helpParameter, t)
}
