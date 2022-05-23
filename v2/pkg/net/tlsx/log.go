// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import "log"

var (
	Log   = func(args ...any) {}
	Logf  = func(format string, args ...any) {}
	Elog  = log.Println
	Elogf = log.Printf
)
