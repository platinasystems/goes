// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "context"

var (
	pathMark int
	pathKey  = &pathMark
)

// PathOf() returns a FIFO of each element appended to the context.
func PathOf(ctx context.Context) []string {
	var l []string
	for v := ctx.Value(pathKey); v != nil; v = ctx.Value(pathKey) {
		p := v.(path)
		l = append(l, p.name)
		ctx = p.Context
	}
	j := len(l)
	for i, s := range l {
		if j -= 1; j <= i {
			break
		}
		l[i] = l[j]
		l[j] = s
	}
	return l
}

// Append a element to context.
func WithPath(ctx context.Context, name string) context.Context {
	return path{ctx, name}
}

type path struct {
	context.Context
	name string
}

func (p path) Value(k interface{}) interface{} {
	if k == pathKey {
		return p
	}
	return p.Context.Value(k)
}
