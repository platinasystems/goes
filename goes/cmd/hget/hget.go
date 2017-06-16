// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hget

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hget"
	Apropos = "get the value of a redis hash field"
	Usage   = "hget KEY FIELD"
)

type Interface interface {
	Apropos() lang.Alt
	Complete(...string) []string
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Complete(args ...string) []string {
	return redis.Complete(args...)
}

func (cmd) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD: missing")
	case 1:
		return fmt.Errorf("FIELD: missing")
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	s, err := redis.Hget(args[0], args[1])
	if err != nil {
		return err
	}
	redis.Fprintln(os.Stdout, s)
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
