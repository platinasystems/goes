// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package redial

import "golang.org/x/sys/unix"

var Errors = []error{
	unix.ECONNREFUSED,
	unix.EHOSTUNREACH,
	unix.ENETUNREACH,
}
