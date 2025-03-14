// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package xsignal

import "golang.org/x/sys/unix"

const (
	Alarm     = unix.SIGALRM
	Terminate = unix.SIGTERM
)
