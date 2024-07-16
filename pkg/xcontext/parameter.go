// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xcontext

import "context"

func AppendParameter[T any](ctx context.Context, key *[]T, value T) context.Context {
	return ContextParameter(ctx, key, append(ParameterValue(ctx, key), value))
}

func ContextParameter[T any](ctx context.Context, key *T, value T) context.Context {
	return context.WithValue(ctx, key, value)
}

// If present, return the context's parameter value; otherwise, return what's
// pointed to by key.
func ParameterValue[T any](ctx context.Context, key *T) T {
	if t := ctx.Value(key); t != nil {
		if v, ok := t.(T); ok {
			return v
		}
	}
	return *key
}
