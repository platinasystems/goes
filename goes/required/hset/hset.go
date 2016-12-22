// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hset

import (
	"fmt"

	"github.com/platinasystems/go/goes/internal/flags"
	"github.com/platinasystems/go/redis"
)

const Name = "hset"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [-q] KEY FIELD VALUE" }

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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "set the string value of a redis hash field",
	}
}

func (cmd) Complete(args ...string) []string {
	return redis.Complete(args...)
}
