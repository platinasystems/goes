// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "golang.org/x/sys/unix"

func InteractiveSysProcAttr() (*unix.SysProcAttr, error) {
	return &unix.SysProcAttr{
		Setsid: true,
		// FIXME?
		// Setctty: true,
		// Ctty:    0,
	}, nil
}
