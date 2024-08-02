// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

//go:build unix && !android && !linux

package xos

import "golang.org/x/sys/unix"

var Select = unix.Select
