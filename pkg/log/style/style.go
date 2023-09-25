// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	plain     = 0
	shortfile = log.Lshortfile
)

type Style struct {
	Level
	*log.Logger
}

type Severity struct{ Errata, Notice Style }

var (
	Mute  = func(args ...any) {}
	Mutef = func(format string, args ...any) {}

	Plain = Severity{
		Style{Errata, log.New(os.Stderr, "", plain)},
		Style{Notice, log.New(os.Stdout, "", plain)},
	}
	ShortFile = Severity{
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

	Recovery = ShortFile.Recovery
)

func (severity Severity) Recovery(suppress ...error) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			for _, ignore := range suppress {
				if errors.Is(err, ignore) {
					return
				}
			}
			depth := 3
			if _, ok := err.(runtime.Error); ok {
				depth = 4
			}
			severity.Errata.Output(depth, err.Error())
		} else {
			severity.Errata.Output(3, fmt.Sprint(r))
		}
	}
}

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
	Helper()
	Error(args ...any)
	Errorf(format string, args ...any)
	FailNow()
	Log(args ...any)
	Logf(format string, args ...any)
}) {
	Error = func(v ...any) {
		t.Helper()
		t.Error(fmt.Sprint(v...))
	}
	Errorf = t.Errorf
	Errorln = t.Error
	Print = func(v ...any) {
		t.Helper()
		t.Log(fmt.Sprint(v...))
	}
	Printf = t.Logf
	Println = t.Log
	Recovery = func(suppress ...error) {
		if r := recover(); r != nil {
			t.Helper()
			if err, ok := r.(error); ok {
				for _, ignore := range suppress {
					if errors.Is(err, ignore) {
						t.Log(err)
						return
					}
				}
				if _, ok := r.(runtime.Error); ok {
					ShortFile.Errata.Output(4, err.Error())
					t.FailNow()
				}
			} else {
				t.Error(r)
			}
		}
	}
}
