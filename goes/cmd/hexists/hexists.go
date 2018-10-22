// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hexists

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct{}

func (Command) String() string { return "hexists" }
func (Command) Usage() string  { return "hexists KEY FIELD" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "determine if the redis hash field exists",
	}
}

func (Command) Complete(args ...string) []string {
	return redis.Complete(args...)
}

func (Command) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD: missing")
	case 1:
		return fmt.Errorf("FIELD: missing")
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	r, err := redis.Connect()
	if err != nil {
		return err
	}
	defer r.Close()
	ret, err := r.Do("HEXISTS", args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(ret)
	return nil
}
