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
	Elog  = style.ShortFile.Errata.Println
	Elogf = style.ShortFile.Errata.Printf
)

func SetVerbosity() {
	verbose := flag.Lookup("verbose")
	if verbose != nil && verbose.Value.String() == "true" {
		Log = style.ShortFile.Notice.Println
		Logf = style.ShortFile.Notice.Printf
	}
}
