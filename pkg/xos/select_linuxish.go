// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

//go:build android || linux

package xos

import "golang.org/x/sys/unix"

// Returns EAGAIN if unix.Select returns “n” == 0.
func Select(
	nfds int,
	rfds, wfds, efds *unix.FdSet,
	tv *unix.Timeval,
) (int, error) {
	n, err := unix.Select(nfds, rfds, wfds, efds, tv)
	if err == nil && n == 0 {
		err = unix.EAGAIN
	}
	return n, err
}
