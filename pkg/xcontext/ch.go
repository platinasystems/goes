// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import (
	. "context"
)

// Return next channeled item unless context is done or channel is closed.
func Next[T any](ctx Context, ch <-chan T) (item T, ok bool) {
	select {
	case <-ctx.Done():
	case item, ok = <-ch:
	}
	return
}

// Returns false if context is done before item is channeled.
func Queue[T any](ctx Context, ch chan<- T, item T) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- item:
		return true
	}
}

// Call function with each channeled item until
// the context is done,
// the channel is closed,
// or the function returns false.
func Range[T any](ctx Context, ch <-chan T, f func(T) bool) {
	for item, ok := Next(ctx, ch); ok; item, ok = Next(ctx, ch) {
		if !f(item) {
			break
		}
	}
}
