// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"os"
	"strings"

	info "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/indent"
)

type ShowMachine string

func (name ShowMachine) String() string { return string(name) }
func (ShowMachine) Usage() string       { return "show machine" }

func (ShowMachine) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print machine name",
	}
}

func (name ShowMachine) Main(...string) error {
	fmt.Println(name)
	return nil
}

type ShowPackages struct{}

func (ShowPackages) String() string { return "packages" }

func (ShowPackages) Usage() string { return "show packages" }

func (ShowPackages) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print machine's package info",
	}
}

func (ShowPackages) Main(...string) error {
	b, err := info.Marshal()
	if err == nil {
		_, err = os.Stdout.Write(b)
	}
	return err
}

func (*Goes) copyright(_ ...string) error {
	printEither("copyright", "license")
	return nil
}

func (*Goes) license(_ ...string) error {
	printEither("license", "copyright")
	return nil
}

func (*Goes) version(_ ...string) error {
	printEither("tag", "version")
	return nil
}

func printEither(alternatives ...string) {
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
