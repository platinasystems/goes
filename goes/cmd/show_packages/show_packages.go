// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_packages

import (
	"fmt"
	"os"
	"strings"

	info "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/indent"
)

const (
	ShowPackagesAproposEnUS = "print package repos info"
	LicenseAproposEnUS      = "print package license(s)"
	VersionAproposEnUS      = "print package version(s)"
	Usage                   = `
	show-packages (default)
	version
	license
	`
)

func New(s string) Command { return Command(s) }

type Command string

func (c Command) Apropos() lang.Alt {
	aproposEnUS := ShowPackagesAproposEnUS
	switch c {
	case "version":
		aproposEnUS = VersionAproposEnUS
	case "license":
		aproposEnUS = LicenseAproposEnUS
	}
	return lang.Alt{
		lang.EnUS: aproposEnUS,
	}
}

func (c Command) String() string { return string(c) }
func (Command) Usage() string    { return Usage }

func (c Command) Main(args ...string) error {
	if len(args) != 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	printEither := func(alternatives ...string) {
		w := indent.New(os.Stdout, "    ")
		for _, m := range info.All() {
			ip := m["importpath"]
			if len(ip) == 0 {
				continue
			}
			for _, alt := range alternatives {
				val := m[alt]
				if len(val) == 0 {
					continue
				}
				fmt.Fprint(w, ip, ": ")
				if strings.Contains(val, "\n") {
					fmt.Fprintln(w, "|")
					indent.Increase(w)
					fmt.Fprintln(w, val)
					indent.Decrease(w)
				} else {
					fmt.Fprintln(w, val)
				}
				break
			}
		}
	}
	switch c {
	case "version":
		printEither("tag", "version")
	case "license":
		printEither("license", "copyright")
	case "copyright":
		printEither("copyright", "license")
	case "show-packages":
		fallthrough
	default:
		b, err := info.Marshal()
		if err == nil {
			_, err = os.Stdout.Write(b)
		}
		return err
	}
	return nil
}
