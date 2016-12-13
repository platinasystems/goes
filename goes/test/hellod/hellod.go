// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hellod

import (
	"fmt"

	"github.com/platinasystems/go/goes"
)

const Name = "hellod"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Kind() goes.Kind { return goes.Daemon }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return Name + " [MESSAGE]..." }

func (cmd) Main(args ...string) error {
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "test daemon info log",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	hellod - test daemon info log

SYNOPSIS
	hellod [MESSAGE]...

DESCRIPTION
	Print the given or default message to klog or syslog.`,
	}
}
