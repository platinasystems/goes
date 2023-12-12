// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

//go:build android || linux

package fdset

import "syscall"

// Returns EAGAIN if syscall.Select returns `n` == 0.
func Select(
	nfds int,
	rfds, wfds, efds *syscall.FdSet,
	tv *syscall.Timeval,
) error {
	n, err := syscall.Select(nfds, rfds, wfds, efds, tv)
	if err == nil && n == 0 {
		err = syscall.EAGAIN
	}
	return err
}
