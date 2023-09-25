// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd || openbsd

package netif

import "syscall"

const SIOCAIFADDR = syscall.SIOCAIFADDR
