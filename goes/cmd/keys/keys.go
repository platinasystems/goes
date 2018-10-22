// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct{}

func (Command) String() string { return "keys" }

func (Command) Usage() string { return "keys [PATTERN]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "find all redis keys matching the given pattern",
	}
}

func (Command) Main(args ...string) error {
	var pattern string
	switch len(args) {
	case 0:
		pattern = ".*"
	case 1:
		pattern = args[0]
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	keys, err := redis.Keys(pattern)
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
