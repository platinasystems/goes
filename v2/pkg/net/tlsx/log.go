// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/log/style"
)

var (
	Log   = style.Muteln
	Logf  = style.Mutef
	Elog  = style.ShortFileStderr.Println
	Elogf = style.ShortFileStderr.Printf
)

func SetVerbosity() {
	verbose := flag.Lookup("verbose")
	if verbose != nil && verbose.Value.String() == "true" {
		Log = style.ShortFileStdout.Println
		Logf = style.ShortFileStdout.Printf
	}
}
