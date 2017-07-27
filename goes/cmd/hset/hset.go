// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hset

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hset"
	Apropos = "set the string value of a redis hash field"
	Usage   = "hset [-q] KEY FIELD VALUE"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Complete(args ...string) []string {
	return redis.Complete(args...)
}

func (Command) Main(args ...string) error {
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
	if !flag.ByName["-q"] {
		fmt.Println(i)
	}
	return nil
}
