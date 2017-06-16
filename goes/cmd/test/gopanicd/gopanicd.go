// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gopanicd

import (
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "go-panicd"
	Apropos = "test daemon error log from go-routine"
	Usage   = "go-panicd [MESSAGE]..."
	Man     = `
DESCRIPTION
	Print the given or default message to klog or syslog followed by
	go-routine trace.`
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
func (cmd) Usage() string     { return Usage + "" }

func (cmd) Main(args ...string) error {
	msg := "---"
	if len(args) > 0 {
		msg = strings.Join(args, " ")
	}
	stop := make(chan error)
	go func() {
		defer func() { stop <- nil }()
		panic(msg)
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
