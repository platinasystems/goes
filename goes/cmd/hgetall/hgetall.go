// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hgetall

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct{}

func (Command) String() string { return "hgetall" }

func (Command) Usage() string { return "hgetall [KEY]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "get all the field values in a redis hash",
	}
}

func (Command) Main(args ...string) error {
	switch len(args) {
	case 0:
		args = []string{redis.DefaultHash}
	case 1:
	default:
		return fmt.Errorf("%v: unexpected; use: `hget %s '%s'`",
			args[1:], args[0], args[1])
	}
	r, err := redis.Connect()
	if err != nil {
		return err
	}
	defer r.Close()
	ret, err := r.Do("HGETALL", args[0])
	if err != nil {
		return err
	}
	list := ret.([]interface{})
	for i := 0; i < len(list); i += 2 {
		fmt.Print(redis.Quotes(string(list[i].([]byte))))
		if list[i+1] != nil {
			fmt.Print(": ")
			fmt.Print(redis.Quotes(
				string(list[i+1].([]byte))))
		}
		fmt.Println()
	}
	return nil
}

func (Command) Complete(args ...string) []string {
	return redis.Complete(args...)
}
