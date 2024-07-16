// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !freebsd

package netrt

func SetFib[FD ~int](sock FD, fib int) error {
	return ErrUnsupported
}
