// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parameter

type Parameter[T any] struct {
	Default T
}

// Prepend first optional parameter value to context.
// If no values are given, use the paramenter default.
func (p *Parameter[T]) With(ctx Context, opt ...T) Context {
	v := p.Default
	if len(opt) > 0 {
		v = opt[0]
	}
	return WithValue(ctx, p, v)
}

// If present, return the context parameter value, otherwise, return
// the parameter default.
func (p *Parameter[T]) Within(ctx Context) T {
	if t := ctx.Value(p); t != nil {
		if v, ok := t.(T); ok {
			return v
		}
	}
	return p.Default
}
