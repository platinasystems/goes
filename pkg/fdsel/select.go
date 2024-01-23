// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

//go:build !android && !linux && !windows && !plan9

package fdset

import "golang.org/x/sys/unix"

var Select = unix.Select
