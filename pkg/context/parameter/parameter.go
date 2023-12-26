// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package parameter

import "context"

func Append[T any](ctx context.Context, key *[]T, value T) context.Context {
	return Context(ctx, key, append(Value(ctx, key), value))
}

func Context[T any](ctx context.Context, key *T, value T) context.Context {
	return context.WithValue(ctx, key, value)
}

// If present, return the context's parameter value; otherwise, return what's
// pointed to by key.
func Value[T any](ctx context.Context, key *T) T {
	if t := ctx.Value(key); t != nil {
		if v, ok := t.(T); ok {
			return v
		}
	}
	return *key
}
