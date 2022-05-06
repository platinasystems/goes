// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "context"

var (
	rootMark int
	rootKey  = &rootMark
)

func RootOf(ctx context.Context) Selection {
	if v := ctx.Value(rootKey); v != nil {
		return v.(Selection)
	}
	return Selection{}
}

func WithRoot(ctx context.Context, m Selection) context.Context {
	return context.WithValue(ctx, rootKey, m)
}
