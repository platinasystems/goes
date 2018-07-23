// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import "flag"

var (
	IsMain = flag.Bool("test.main", false, "run main() instead of test(s)")
	VV     = flag.Bool("test.vv", false, "log test.Program output")
	VVV    = flag.Bool("test.vvv", false, "log test.Program execution")
)
