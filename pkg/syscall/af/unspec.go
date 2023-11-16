// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || darwin || freebsd || ios || linux || netbsd || openbsd || windows

package af

import "syscall"

const UNSPEC = syscall.AF_UNSPEC
