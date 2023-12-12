// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

//go:build !android && !linux && !windows && !plan9

package fdset

import "syscall"

var Select = syscall.Select
