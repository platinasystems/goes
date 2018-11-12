// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
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

func (g *Goes) copyright(args ...string) error {
	return g.license(args...)
}

func (g *Goes) license(args ...string) error {
	type licenser interface {
		License() error
	}
	f := func() error {
		fmt.Println(License)
		return nil
	}
	if len(args) > 0 {
		if v, found := g.ByName[args[0]]; found {
			if method, found := v.(licenser); found {
				f = method.License
			}
		}
	}
	return f()
}

func (g *Goes) patents(args ...string) error {
	type patentser interface {
		Patents() error
	}
	f := func() error {
		fmt.Println(Patents)
		return nil
	}
	if len(args) > 0 {
		if v, found := g.ByName[args[0]]; found {
			if method, found := v.(patentser); found {
				f = method.Patents
			}
		}
	}
	return f()
}

func (g *Goes) version(args ...string) error {
	type versioner interface {
		Version() error
	}
	f := func() error {
		fmt.Println(Version)
		return nil
	}
	if len(args) > 0 {
		if v, found := g.ByName[args[0]]; found {
			if method, found := v.(versioner); found {
				f = method.Version
			}
		}
	}
	return f()
}
