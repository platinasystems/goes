// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"fmt"
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
	Base = program.Base()

	notice = Base + ":"
	errata = Base + ":error:"

	Mute  = func(args ...any) {}
	Mutef = func(format string, args ...any) {}

	Plain = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, errata+" ", plain)},
		Style{Notice, log.New(os.Stdout, notice+" ", plain)},
	}
	ShortFile = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, errata, shortfile)},
		Style{Notice, log.New(os.Stdout, notice, shortfile)},
	}

	Error   = ShortFile.Errata.Print
	Errorf  = ShortFile.Errata.Printf
	Errorln = ShortFile.Errata.Println
	Fatal   = Plain.Errata.Fatal
	Fatalf  = Plain.Errata.Fatalf
	Fatalln = Plain.Errata.Fatalln
	Print   = Mute
	Printf  = Mutef
	Println = Mute

	Quiet, Verbose *bool
)

// Log Plain and ShortFile messages to System logger instead of Std{out|err}.
func System() {
	Plain.Errata.System()
	Plain.Notice.System()
	ShortFile.Errata.System()
	ShortFile.Notice.System()
}

func Verbosity() {
	if Quiet != nil && *Quiet {
		Error = Mute
		Errorf = Mutef
		Errorln = Mute
	}
	if Verbose != nil && *Verbose {
		Print = ShortFile.Notice.Print
		Printf = ShortFile.Notice.Printf
		Println = ShortFile.Notice.Println
	}
}

func Test(t interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
}) {
	Error = func(v ...any) { t.Error(fmt.Sprint(v...)) }
	Errorf = t.Errorf
	Errorln = t.Error
	Print = func(v ...any) { t.Log(fmt.Sprint(v...)) }
	Printf = t.Logf
	Println = t.Log
}
