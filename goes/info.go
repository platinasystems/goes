// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/lang"
)

var Info struct {
	Licenses, Patents, Versions func() map[string]string
}

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

func (g *Goes) copyright(_ ...string) error {
	return g.license()
}

func (*Goes) license(_ ...string) error {
	marshal(Info.Licenses)
	return nil
}

func (*Goes) patents(_ ...string) error {
	marshal(Info.Patents)
	return nil
}

func (*Goes) version(_ ...string) error {
	marshal(Info.Versions)
	return nil
}

func marshal(f func() map[string]string) {
	var sep string
	if f == nil {
		return
	}
	m := f()
	for k, v := range m {
		s := strings.TrimSpace(v)
		if len(m) == 1 {
			fmt.Println(s)
		} else if !strings.ContainsRune(s, '\n') {
			fmt.Print(sep, k, ": ", s, "\n")
		} else {
			fmt.Print(sep, k, ": |\n")
			for _, l := range strings.Split(s, "\n") {
				fmt.Print("  ", l, "\n")
			}
			sep = "\n"
		}
	}
}
