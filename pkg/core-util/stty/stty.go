// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package stty

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"golang.org/x/term"
)

type namedVchar struct {
	name  string
	vchar uint
}

type namedFlag struct {
	name string
	flag uint64
}

const SttyUsage = `
usage: {{.Name}} [flags] [setting(s)]
Change and print terminal line settings.
` + Settings

var (
	Stty_a,
	Stty_g bool
	Stty_f string
)

var SttyFlags = xflag.Labels{
	{"a", "Display all settings.", &Stty_a},
	{"f", "Usage named TTY instead of stdin", &Stty_f},
	{"g", "Display restore settings.", &Stty_g},
}

var tty = os.Stdin

func Stty(ctx context.Context, args []string) error {
	xflag.TemplateUsage(SttyUsage)
	err := SttyFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	if len(Stty_f) > 0 {
		var err error
		tty, err = os.Open(Stty_f)
		if err != nil {
			return err
		}
		defer tty.Close()
	}
	if !term.IsTerminal(int(tty.Fd())) {
		return fmt.Errorf("%s: isn't TTY")
	}
	if err = get(); err != nil {
		return err
	}
	if Stty_g {
		printRestoration()
	}
	if args = flag.Args(); len(args) != 0 {
		if err = set(args); err != nil {
			return err
		}
	}
	if !Stty_g {
		err = printSettings()
	}
	return err
}

func printSettings() error {
	fmt.Print("speed ", baud(), " baud")
	if Stty_a {
		w, h, err := term.GetSize(int(tty.Fd()))
		if err != nil {
			return err
		}
		fmt.Print("; rows ", h, "; columns ", w)
	}
	fmt.Println(";")
	printFlags("lflags:", getLflag(), namedLFlags)
	printFlags("iflags:", getIflag(), namedIFlags)
	printFlags("oflags:", getOflag(), namedOFlags)
	printFlags("cflags:", getCflag(), namedCFlags)
	printChars("cchars:", getCc())
	return nil
}

func printChars(heading string, cc []uint8) {
	if !Stty_a {
		return
	}
	fmt.Print(heading)
	n := len(heading)
	indent := strings.Repeat(" ", n)
	s := new(strings.Builder)
	for _, entry := range namedVchars {
		if entry.vchar >= uint(len(cc)) {
			break
		}
		r := rune(cc[entry.vchar])
		s.Reset()
		fmt.Fprint(s, " ", entry.name, " = ")
		if r == 0 || r == 255 {
			fmt.Fprint(s, "<undef>")
		} else if r < ' ' {
			fmt.Fprintf(s, "^%c", r+'A'-1)
		} else if r == 127 {
			fmt.Fprint(s, "^?")
		} else {
			fmt.Fprint(s, r)
		}
		fmt.Fprint(s, ";")
		n += s.Len()
		if n >= 80 {
			fmt.Print("\n", indent)
			n = len(heading)
		}
		fmt.Print(s)
	}
	fmt.Println()
}

func printFlags(heading string, flags uint64, namedFlags []namedFlag) {
	indent := strings.Repeat(" ", len(heading))
	if flags == 0 && !Stty_a {
		return
	}
	fmt.Print(heading)
	n := len(heading)
	if heading == "cflags:" {
		csn, _ := fmt.Print(" ", getCSize())
		n += csn
	}
	s := new(strings.Builder)
	for _, entry := range namedFlags {
		s.Reset()
		if flags&entry.flag == entry.flag {
			fmt.Fprint(s, " ", entry.name)
		} else if Stty_a {
			fmt.Fprint(s, " -", entry.name)
		} else {
			continue
		}
		n += s.Len()
		if n >= 80 {
			fmt.Print("\n", indent)
			n = len(heading)
		}
		fmt.Print(s)
	}
	fmt.Println()
}
