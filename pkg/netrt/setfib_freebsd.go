// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "golang.org/x/sys/unix"

func SetFib[FD ~int](fd FD, fib int) error {
	return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SETFIB, fib)
}
