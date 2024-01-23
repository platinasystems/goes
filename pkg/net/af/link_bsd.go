// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || ios || netbsd || openbsd

package af

import "golang.org/x/sys/unix"

const LINK = unix.AF_LINK
