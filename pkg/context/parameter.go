// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package context

// A Parameter provides a key, type, default value and methods to get and
// put values in context.
type Parameter[T any] struct {
	Default T
}

func NewParameter[T any](defval T) *Parameter[T] {
	return &Parameter[T]{defval}
}

// If present, return parameter value in context, otherwise, return the
// parameter default.
func (p *Parameter[T]) In(ctx Context) T {
	if t := ctx.Value(p); t != nil {
		if v, ok := t.(T); ok {
			return v
		}
	}
	return p.Default
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
