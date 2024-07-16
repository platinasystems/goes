// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !unix

package xos

import . "os"

var Termination = []Signal{Interrupt}
