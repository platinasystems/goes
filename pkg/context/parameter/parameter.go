// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parameter

import (
	"context"
)

type Key[T any] struct{ _ T }

// Prepend key'd value to context.
func (k *Key[T]) With(ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, k, v)
}

// Return key'd value within context or its zero value if not present.
func (k *Key[T]) Value(ctx context.Context) T {
	v, _ := ctx.Value(k).(T)
	return v
}
