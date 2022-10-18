// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"log"
	"os"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const (
	plain     = 0
	shortfile = log.Lshortfile
)

type Style struct {
	Level
	*log.Logger
}

var (
	Base = program.Base.Value()

	notice = Base + ":"
	errata = Base + ":error:"

	Mute   = func(args ...any) {}
	Muteln = func(args ...any) {}
	Mutef  = func(format string, args ...any) {}

	Plain = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, notice+" ", plain)},
		Style{Notice, log.New(os.Stdout, errata+" ", plain)},
	}
	ShortFile = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, notice, shortfile)},
		Style{Notice, log.New(os.Stdout, errata, shortfile)},
	}
)

// Log Plain and ShortFile messages to System logger instead of Std{out|err}.
func System() {
	Plain.Errata.System()
	Plain.Notice.System()
	ShortFile.Errata.System()
	ShortFile.Notice.System()
}
