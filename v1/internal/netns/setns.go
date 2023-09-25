// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux,!amd64

package netns

import "syscall"

const SYS_SETNS = syscall.SYS_SETNS
