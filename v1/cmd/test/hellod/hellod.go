// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hellod

import (
	"fmt"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

const (
	Name    = "hellod"
	Apropos = "test daemon info log"
	Usage   = "hellod [MESSAGE]..."
	Man     = `
DESCRIPTION
	Print the given or default message to klog or syslog.`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) Kind() cmd.Kind    { return cmd.Daemon }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	iargs := []interface{}{"hello", "world"}
	if len(args) > 0 {
		iargs = make([]interface{}, 0, len(args))
		for _, arg := range args {
			iargs = append(iargs, arg)
		}
	}
	fmt.Println(iargs...)
	return nil
}
