// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux || freebsd || netbsd || openbsd || darwin || windows

package redial

import "syscall"

var Errors = []error{
	syscall.ECONNREFUSED,
	syscall.EHOSTUNREACH,
	syscall.ENETUNREACH,
}
