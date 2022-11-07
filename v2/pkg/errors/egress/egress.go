// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package egress

// Returns a closure that calls f() before returning an empty T and err.
func New[T any](f func()) func(error) (T, error) {
	return func(err error) (T, error) {
		f()
		var t T
		return t, err
	}
}
