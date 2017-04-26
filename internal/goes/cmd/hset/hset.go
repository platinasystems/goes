// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hset

import (
	"fmt"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hset"
	Apropos = "set the string value of a redis hash field"
	Usage   = "hset [-q] KEY FIELD VALUE"
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
	flag, args := flags.New(args, "-q")
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD VALUE: missing")
	case 1:
		return fmt.Errorf("FIELD VALUE: missing")
	case 2:
		return fmt.Errorf("VALUE: missing")
	case 3:
	default:
		return fmt.Errorf("%v: unexpected", args[3:])
	}
	i, err := redis.Hset(args[0], args[1], args[2])
	if err != nil {
		return err
	}
	if !flag["-q"] {
		fmt.Println(i)
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
