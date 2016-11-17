// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package lrange

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/redis/internal"
	"github.com/platinasystems/go/redis"
)

const Name = "lrange"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " KEY START STOP" }

func (cmd) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY START STOP: missing")
	case 1:
		return fmt.Errorf("START STOP: missing")
	case 2:
		return fmt.Errorf("STOP: missing")
	case 3:
	default:
		return fmt.Errorf("%v: unexpected", args[3:])
	}
	var start, stop int
	if _, err := fmt.Sscan(args[1], &start); err != nil {
		return err
	}
	if _, err := fmt.Sscan(args[2], &stop); err != nil {
		return err
	}
	keys, err := redis.Lrange(args[0], start, stop)
	if err != nil {
		return err
	}
	for _, s := range keys {
		internal.Fprintln(os.Stdout, s)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get a range of elements from a redis list",
	}
}

func (cmd) Complete(args ...string) []string {
	return internal.Complete(args...)
}
