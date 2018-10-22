// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hkeys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct{}

func (Command) String() string { return "hkeys" }

func (Command) Usage() string { return "hkeys KEY" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "get all the fields in a redis hash",
	}
}

func (Command) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY: missing")
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	keys, err := redis.Hkeys(args[0])
	if err != nil {
		return err
	}
	for _, s := range keys {
		redis.Fprintln(os.Stdout, s)
	}
	return nil
}

func (Command) Complete(args ...string) []string {
	return redis.Complete(args...)
}
