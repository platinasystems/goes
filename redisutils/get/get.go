// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package get

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

const Name = "get"

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
	s, err := redis.Get(args[0])
	if err != nil {
		return err
	}
	internal.Fprintln(os.Stdout, s)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get the value of a redis key",
	}
}

func (cmd) Complete(args ...string) []string {
	return internal.Complete(args...)
}
