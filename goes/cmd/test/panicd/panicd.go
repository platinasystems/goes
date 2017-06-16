// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package panicd

import (
	"strings"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "panicd"
	Apropos = "test daemon error log"
	Usage   = "panicd [MESSAGE]..."
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
	msg := "---"
	if len(args) > 0 {
		msg = strings.Join(args, " ")
	}
	panic(msg)
	return nil
}
