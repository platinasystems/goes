// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gohellod

import (
	"fmt"

	"github.com/platinasystems/go/internal/goes"
)

const Name = "go-hellod"

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
	stop := make(chan error)
	go func() {
		defer func() { stop <- nil }()
		fmt.Println(iargs...)
	}()
	return <-stop
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "test daemon info log from go-routine",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	go-panicd - test daemon info log from go-routine

SYNOPSIS
	go-panicd [MESSAGE]...

DESCRIPTION
	Print the given or default message to klog or syslog.`,
	}
}
