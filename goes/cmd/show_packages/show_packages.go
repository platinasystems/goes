// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_packages

import (
	"fmt"
	"os"

	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show-packages"
	Apropos = "print package repos info"
	Usage   = "version"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Kind() cmd.Kind    { return cmd.DontFork }

func (Command) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	_, err := WriteTo(os.Stdout)
	return err
}

func (Command) String() string { return Name }
func (Command) Usage() string  { return Usage }
