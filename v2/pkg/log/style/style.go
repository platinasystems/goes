// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	Mute  = func(args ...any) {}
	Mutef = func(format string, args ...any) {}

	Plain = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, "", plain)},
		Style{Notice, log.New(os.Stdout, "", plain)},
	}
	ShortFile = struct{ Errata, Notice Style }{
		Style{Errata, log.New(os.Stderr, "", shortfile)},
		Style{Notice, log.New(os.Stdout, "", shortfile)},
	}

	// Errors may be muted with the quiet flag.
	Error   = ShortFile.Errata.Print
	Errorf  = ShortFile.Errata.Printf
	Errorln = ShortFile.Errata.Println
	// Fatal messages are never muted.
	Fatal   = Plain.Errata.Fatal
	Fatalf  = Plain.Errata.Fatalf
	Fatalln = Plain.Errata.Fatalln
	// Notes may be muted with the quiet flag.
	Note   = ShortFile.Notice.Print
	Notef  = ShortFile.Notice.Printf
	Noteln = ShortFile.Notice.Println
	// Prints may be unmuted with the verbose flag.
	Print   = Mute
	Printf  = Mutef
	Println = Mute

	Quiet   = flag.Bool("quiet", false, "Suppress errata.")
	Verbose = flag.Bool("verbose", false, "Print notices.")
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
		Note = Mute
		Notef = Mutef
		Noteln = Mute
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
