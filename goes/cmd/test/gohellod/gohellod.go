// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gohellod

import (
	"fmt"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "go-hellod"
	Apropos = "test daemon info log from go-routine"
	Usage   = "go-hellod [MESSAGE]..."
	Man     = `
DESCRIPTION
	Print the given or default message to klog or syslog.`
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) Kind() goes.Kind   { return goes.Daemon }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

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

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
