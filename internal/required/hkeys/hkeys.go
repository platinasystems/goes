// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hkeys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/redis"
)

const Name = "hkeys"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " KEY" }

func (cmd) Main(args ...string) error {
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get all the fields in a redis hash",
	}
}

func (cmd) Complete(args ...string) []string {
	return redis.Complete(args...)
}
